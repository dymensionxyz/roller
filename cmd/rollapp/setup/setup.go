package setup

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	cosmossdkmath "cosmossdk.io/math"
	cosmossdktypes "github.com/cosmos/cosmos-sdk/types"
	dymensionseqtypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	"github.com/dymensionxyz/roller/data_layer/celestia"
	celestialightclient "github.com/dymensionxyz/roller/data_layer/celestia/lightclient"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/denom"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/genesis"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/rollapp"
	"github.com/dymensionxyz/roller/utils/rollapp/iro"
	"github.com/dymensionxyz/roller/utils/roller"
	"github.com/dymensionxyz/roller/utils/sequencer"
)

// TODO: Test sequencing on 35-C and update the price
var OneDaySequencePrice = big.NewInt(1)

var LogPath string

// nolint:gocyclo
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Setup a RollApp node.",
		Long:  ``,
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			nodeTypes := []string{"sequencer", "fullnode"}
			fullNodeTypes := []string{"rpc", "archive"}

			nodeTypeFromFlag, _ := cmd.Flags().GetString("node-type")
			fullNodeTypeFromFlag, _ := cmd.Flags().GetString("full-node-type")
			shouldUseDefaultRpcEndpoint, _ := cmd.Flags().GetBool("use-default-rpc-endpoint")

			err := initconfig.AddFlags(cmd)
			if err != nil {
				pterm.Error.Println("failed to add flags")
				return
			}

			home, err := filesystem.ExpandHomePath(
				cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String(),
			)
			if err != nil {
				pterm.Error.Println("failed to expand home directory")
				return
			}

			localRollerConfig, err := roller.LoadConfig(home)
			if err != nil {
				pterm.Error.Println("failed to load roller config file", err)
				return
			}

			if localRollerConfig.Environment == "mock" {
				pterm.Error.Println("setup is not required for mock backend")
				pterm.Info.Printf(
					"run %s instead to run the rollapp\n",
					pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
						Sprintf("roller rollapp start"),
				)
				return
			}

			if !shouldUseDefaultRpcEndpoint {
				localRollerConfig = config.PromptCustomHubEndpoint(localRollerConfig)
			}

			rollappConfig, err := rollapp.PopulateRollerConfigWithRaMetadataFromChain(
				home,
				localRollerConfig.RollappID,
				localRollerConfig.HubData,
			)
			errorhandling.PrettifyErrorIfExists(err)

			raResponse, err := rollapp.GetMetadataFromChain(
				localRollerConfig.RollappID,
				localRollerConfig.HubData,
			)
			if err != nil {
				pterm.Error.Println("failed to fetch rollapp information from hub: ", err)
				return
			}

			ok := iro.IsTokenGraduates(raResponse.Rollapp.RollappId, localRollerConfig.HubData)
			if !ok {
				pterm.Error.Println("the token has not yet graduated")
			}

			bp, err := rollapp.ExtractBech32PrefixFromBinary(
				strings.ToLower(raResponse.Rollapp.VmType),
			)
			if err != nil {
				pterm.Error.Println("failed to extract bech32 prefix from binary", err)
			}

			if raResponse.Rollapp.GenesisInfo.Bech32Prefix != bp {
				pterm.Error.Printf(
					"rollapp bech32 prefix does not match, want: %s, have: %s\n",
					raResponse.Rollapp.GenesisInfo.Bech32Prefix,
					bp,
				)
				// TODO: reinstall rollapp binary with the right bech32 prefix
				return
			}

			var nodeType string
			if slices.Contains(nodeTypes, nodeTypeFromFlag) {
				nodeType = nodeTypeFromFlag
			} else {
				nodeType, _ = pterm.DefaultInteractiveSelect.
					WithDefaultText("select the node type you want to run").
					WithOptions(nodeTypes).
					Show()
			}

			rollerConfigFilePath := filepath.Join(home, consts.RollerConfigFileName)
			err = tomlconfig.UpdateFieldInFile(rollerConfigFilePath, "node_type", nodeType)
			if err != nil {
				pterm.Error.Println("failed to update node type in roller config file: ", err)
				return
			}

			switch nodeType {
			case "sequencer":
				canRegister, err := sequencer.CanSequencerBeRegisteredForRollapp(
					raResponse.Rollapp.RollappId,
					localRollerConfig.HubData,
				)
				if err != nil {
					pterm.Error.Println(
						"failed to check whether a sequencer can be registered for rollapp: ",
						err,
					)
					return
				}

				if !canRegister {
					pterm.Error.Println("rollapp is not ready to register a sequencer")
					return
				}

				if !raResponse.Rollapp.GenesisInfo.Sealed {
					gvSpinner, err := pterm.DefaultSpinner.Start(
						"validating genesis (this can take several minutes for large genesis files)",
					)
					if err != nil {
						pterm.Error.Println("failed to validate genesis: ", err)
					}
					err = genesis.ValidateGenesis(
						localRollerConfig,
						localRollerConfig.RollappID,
						localRollerConfig.HubData,
					)
					// nolint:errcheck
					gvSpinner.Stop()
					if err != nil {
						pterm.Error.Println("failed to validate genesis: ", err)
						return
					}
					gvSpinner.Success("genesis successfully validated")
					fmt.Println()
				}

				pterm.Info.Println("getting the existing sequencer address ")
				hubSeqKC := keys.KeyConfig{
					Dir:            consts.ConfigDirName.HubKeys,
					ID:             consts.KeysIds.HubSequencer,
					ChainBinary:    consts.Executables.Dymension,
					Type:           consts.SDK_ROLLAPP,
					KeyringBackend: localRollerConfig.KeyringBackend,
				}
				seqAddrInfo, err := hubSeqKC.Info(rollappConfig.Home)
				if err != nil {
					pterm.Error.Println("failed to get address info: ", err)
					return
				}
				seqAddrInfo.Address = strings.TrimSpace(seqAddrInfo.Address)

				// check whether the address is registered as sequencer
				pterm.Info.Printf(
					"checking whether sequencer is already registered for %s\n",
					rollappConfig.RollappID,
				)

				seq, err := sequencer.RegisteredRollappSequencersOnHub(
					rollappConfig.RollappID,
					rollappConfig.HubData,
				)
				if err != nil {
					pterm.Error.Println("failed to retrieve registered sequencers: ", err)
				}

				isSequencerRegistered := sequencer.IsRegisteredAsSequencer(
					seq.Sequencers,
					seqAddrInfo.Address,
				)

				if !isSequencerRegistered {
					minBond, err := sequencer.GetMinSequencerBondInBaseDenom(
						rollappConfig.RollappID,
						rollappConfig.HubData,
					)
					if err != nil {
						pterm.Error.Println("failed to get min bond: ", err)
						return
					}

					var bondAmount cosmossdktypes.Coin
					bondAmount.Denom = consts.Denoms.Hub

					floatDenomRepresentation := displayRegularDenom(*minBond, 18)
					displayDenom := fmt.Sprintf(
						"%s%s",
						floatDenomRepresentation,
						consts.Denoms.Hub[1:],
					)

					var desiredBond cosmossdktypes.Coin
					desiredBondAmount, _ := pterm.DefaultInteractiveTextInput.WithDefaultText(
						fmt.Sprintf(
							"what is your desired bond amount? ( min: %s ) press enter to proceed with %s",
							displayDenom,
							displayDenom,
						),
					).WithDefaultValue(floatDenomRepresentation).Show()

					if strings.TrimSpace(desiredBondAmount) == "" {
						desiredBond = *minBond
					} else {
						f, _ := new(big.Float).SetString(desiredBondAmount)

						// Multiply by 10^18
						multiplier := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
						f.Mul(f, multiplier)

						// Convert to big.Int
						i, _ := f.Int(nil)

						desiredBond.Denom = consts.Denoms.Hub
						desiredBond.Amount = cosmossdkmath.NewIntFromBigInt(i)
					}

					pterm.Info.Println("getting the existing sequencer address balance")
					balance, err := keys.QueryBalance(
						keys.ChainQueryConfig{
							Denom:  consts.Denoms.Hub,
							RPC:    rollappConfig.HubData.RpcUrl,
							Binary: consts.Executables.Dymension,
						}, seqAddrInfo.Address,
					)
					if err != nil {
						pterm.Error.Println("failed to get address balance: ", err)
						return
					}

					// Initialize necessaryBalance with the desiredBond amount
					necessaryBalance := desiredBond.Amount
					// Add the default transaction fee
					necessaryBalance = necessaryBalance.Add(
						cosmossdkmath.NewInt(consts.DefaultTxFee),
					)

					blnc, _ := denom.BaseDenomToDenom(*balance, 18)
					oneDym, _ := cosmossdkmath.NewIntFromString("1000000000000000000")

					nb := cosmossdktypes.Coin{
						Denom:  consts.Denoms.Hub,
						Amount: necessaryBalance.Add(oneDym),
					}
					necBlnc, _ := denom.BaseDenomToDenom(nb, 18)

					pterm.Info.Printf(
						"current balance: %s (%s)\nnecessary balance: %s (%s)\n",
						balance.String(),
						blnc.String(),
						fmt.Sprintf("%s%s", necessaryBalance.String(), consts.Denoms.Hub),
						necBlnc.String(),
					)

					isAddrFunded := balance.Amount.GTE(necessaryBalance)
					if !isAddrFunded {
						pterm.DefaultSection.WithIndentCharacter("🔔").
							Println("Please fund the addresses below to register and run the sequencer.")
						seqAddrInfo.Print(keys.WithName())
						proceed, _ := pterm.DefaultInteractiveConfirm.WithDefaultValue(false).
							WithDefaultText(
								"press 'y' when the wallets are funded",
							).Show()

						if !proceed {
							return
						}
					}

					err = populateSequencerMetadata(*rollappConfig)
					if err != nil {
						pterm.Error.Println("failed to populate sequencer metadata: ", err)
						return
					}

					balance, err = keys.QueryBalance(
						keys.ChainQueryConfig{
							Denom:  consts.Denoms.Hub,
							RPC:    rollappConfig.HubData.RpcUrl,
							Binary: consts.Executables.Dymension,
						}, seqAddrInfo.Address,
					)
					if err != nil {
						pterm.Error.Println("failed to get address balance: ", err)
						return
					}
					blnc, _ = denom.BaseDenomToDenom(*balance, 18)

					pterm.Info.Printf(
						"current balance: %s (%s)\nnecessary balance: %s (%s)\n",
						balance.String(),
						blnc.String(),
						fmt.Sprintf("%s%s", necessaryBalance.String(), consts.Denoms.Hub),
						necBlnc.String(),
					)

					// check whether balance is bigger or equal to the necessaryBalance
					isAddrFunded = balance.Amount.GTE(necessaryBalance)
					if !isAddrFunded {
						pterm.DefaultSection.WithIndentCharacter("🔔").
							Println("Please fund the addresses below to register and run the sequencer.")
						seqAddrInfo.Print(keys.WithName())
						proceed, _ := pterm.DefaultInteractiveConfirm.WithDefaultValue(false).
							WithDefaultText(
								"press 'y' when funded",
							).Show()

						if !proceed {
							return
						}
					}

					err = sequencer.Register(*rollappConfig, desiredBond)
					if err != nil {
						pterm.Error.Println("failed to register sequencer: ", err)
						return
					}
					pterm.Info.Printf(
						"%s ( %s ) is registered as a sequencer for %s\n",
						seqAddrInfo.Name,
						seqAddrInfo.Address,
						rollappConfig.RollappID,
					)
				} else {
					pterm.Info.Printf(
						"%s ( %s ) is registered as a sequencer for %s\n",
						seqAddrInfo.Name,
						seqAddrInfo.Address,
						rollappConfig.RollappID,
					)
				}

			case "fullnode":
				pterm.Info.Println("retrieving the latest available snapshot")
				si, err := sequencer.GetLatestSnapshot(
					rollappConfig.RollappID,
					rollappConfig.HubData,
				)
				if err != nil {
					pterm.Error.Println("failed to retrieve latest snapshot")
				}

				if si == nil {
					pterm.Warning.Printf(
						"no snapshots were found for %s, the node will sync from genesis block\n",
						rollappConfig.RollappID,
					)
				} else {
					fmt.Printf(
						"found a snapshot for height %s\nchecksum: %s\nurl: %s\n",
						si.Height,
						si.Checksum,
						si.SnapshotUrl,
					)

					// approve the data directory deletion before downloading the snapshot
					rollappDirPath := filepath.Join(home, consts.ConfigDirName.Rollapp)

					dataDir := filepath.Join(rollappDirPath, "data")
					if fi, err := os.Stat(dataDir); err == nil && fi.IsDir() {
						dataDirNotEmpty, err := filesystem.DirNotEmpty(dataDir)
						if err != nil {
							pterm.Error.Printf("failed to check if data directory is empty: %v\n", err)
							os.Exit(1)
						}

						var replaceExistingData bool
						if dataDirNotEmpty {
							pterm.Warning.Println("the ~/.roller/rollapp/data directory is not empty.")
							replaceExistingData, _ = pterm.DefaultInteractiveConfirm.Show(
								"Do you want to replace its contents?",
							)
							if !replaceExistingData {
								pterm.Info.Println(
									"operation cancelled, node will be synced from genesis block ",
								)
							}
						}

						if !dataDirNotEmpty || replaceExistingData {
							// TODO: this should be a util "RecreateDir"
							err = os.RemoveAll(dataDir)
							if err != nil {
								pterm.Error.Printf("failed to remove %s dir: %v", dataDir, err)
								return
							}

							err = os.MkdirAll(dataDir, 0o755)
							if err != nil {
								pterm.Error.Printf("failed to create %s: %v\n", dataDir, err)
								return
							}

							tmpDir, err := os.MkdirTemp("", "download-*")
							if err != nil {
								fmt.Printf("Error creating temp directory: %v\n", err)
								return
							}

							// Print the path of the temporary directory
							fmt.Printf("Temporary directory created: %s\n", tmpDir)

							// The directory will be deleted when the program exits
							// nolint:errcheck,gosec
							defer os.RemoveAll(tmpDir)
							archivePath := filepath.Join(tmpDir, "backup.tar.gz")
							spinner, _ := pterm.DefaultSpinner.Start("downloading file...")
							downloadedFileHash, err := filesystem.DownloadAndSaveArchive(
								si.SnapshotUrl,
								archivePath,
							)
							if err != nil {
								spinner.Fail(fmt.Sprintf("error downloading file: %v", err))
								os.Exit(1)
							}
							spinner.Success("file downloaded successfully")

							// compare the checksum
							if downloadedFileHash != si.Checksum {
								pterm.Error.Printf(
									"snapshot archive checksum mismatch, have: %s, want: %s.",
									downloadedFileHash,
									si.Checksum,
								)

								return
							}

							err = filesystem.ExtractTarGz(archivePath, filepath.Join(rollappDirPath))
							if err != nil {
								pterm.Error.Println("failed to extract snapshot: ", err)
								return
							}
						}
					}
				}

				// look for p2p bootstrap nodes, if there are no nodes available, the rollapp
				// defaults to syncing only from the DA
				peers, err := sequencer.GetAllP2pPeers(
					rollappConfig.RollappID,
					rollappConfig.HubData,
				)
				if err != nil {
					pterm.Error.Println("failed to retrieve p2p peers ")
				}

				if len(peers) == 0 {
					pterm.Warning.Println(
						"none of the sequencers provide p2p seed nodes this node will sync only from DA",
					)
				} else {
					peers := strings.Join(peers, ",")
					fieldsToUpdate := map[string]any{
						"p2p_bootstrap_nodes":  peers,
						"p2p_persistent_nodes": peers,
					}
					dymintFilePath := sequencer.GetDymintFilePath(rollappConfig.Home)

					err = tomlconfig.UpdateFieldsInFile(dymintFilePath, fieldsToUpdate)
					if err != nil {
						pterm.Warning.Println("failed to add p2p peers: ", err)
					}
				}
			}

			/* ------------------------ Initialize DA ------------------------ */

			var addresses []keys.KeyInfo
			// Generalize DA initialization logic
			switch localRollerConfig.DA.Backend {
			case consts.Celestia:
				// Initialize Celestia light client
				daKeyInfo, err := celestialightclient.Initialize(
					localRollerConfig.HubData.Environment,
					localRollerConfig,
				)
				if err != nil {
					pterm.Error.Println("failed to initialize Celestia light client: %w", err)
					return
				}

				// Append DA account address if available
				if daKeyInfo != nil {
					addresses = append(addresses, *daKeyInfo)
				}

			case consts.Avail:
				// Initialize DAManager for Avail
				damanager := datalayer.NewDAManager(
					consts.Avail,
					home,
					localRollerConfig.KeyringBackend,
					localRollerConfig.NodeType,
				)

				// Retrieve DA account address
				daAddress, err := damanager.GetDAAccountAddress()
				if err != nil {
					pterm.Error.Println("failed to get Avail account address: %w", err)
					return
				}

				// Append DA account address if available
				if daAddress != nil {
					addresses = append(addresses, keys.KeyInfo{
						Name:    damanager.GetKeyName(),
						Address: daAddress.Address,
					})
				}
			case consts.LoadNetwork:
				// Initialize DAManager for LoadNetwork
				damanager := datalayer.NewDAManager(
					consts.LoadNetwork,
					home,
					localRollerConfig.KeyringBackend,
					localRollerConfig.NodeType,
				)

				// Retrieve DA account address
				daAddress, err := damanager.GetDAAccountAddress()
				if err != nil {
					pterm.Error.Println("failed to get LoadNetwork account address: %w", err)
					return
				}

				// Append DA account address if available
				if daAddress != nil {
					addresses = append(addresses, keys.KeyInfo{
						Name:    damanager.GetKeyName(),
						Address: daAddress.Address,
					})
				}
			case consts.Bnb:
				// Initialize DAManager for Bnb
				damanager := datalayer.NewDAManager(
					consts.Bnb,
					home,
					localRollerConfig.KeyringBackend,
					localRollerConfig.NodeType,
				)

				// Retrieve DA account address
				daAddress, err := damanager.GetDAAccountAddress()
				if err != nil {
					pterm.Error.Println("failed to get Bnb account address: %w", err)
					return
				}

				// Append DA account address if available
				if daAddress != nil {
					addresses = append(addresses, keys.KeyInfo{
						Name:    damanager.GetKeyName(),
						Address: daAddress.Address,
					})
				}
			case consts.Sui:
				// Initialize DAManager for Sui
				damanager := datalayer.NewDAManager(
					consts.Sui,
					home,
					localRollerConfig.KeyringBackend,
					localRollerConfig.NodeType,
				)

				// Retrieve DA account address
				daAddress, err := damanager.GetDAAccountAddress()
				if err != nil {
					pterm.Error.Println("failed to get Sui account address: %w", err)
					return
				}

				// Append DA account address if available
				if daAddress != nil {
					addresses = append(addresses, keys.KeyInfo{
						Name:    damanager.GetKeyName(),
						Address: daAddress.Address,
					})
				}
			case consts.Aptos:
				// Initialize DAManager for Aptos
				damanager := datalayer.NewDAManager(
					consts.Aptos,
					home,
					localRollerConfig.KeyringBackend,
					localRollerConfig.NodeType,
				)

				// Retrieve DA account address
				daAddress, err := damanager.GetDAAccountAddress()
				if err != nil {
					pterm.Error.Println("failed to get Aptos account address: %w", err)
					return
				}

				// Append DA account address if available
				if daAddress != nil {
					addresses = append(addresses, keys.KeyInfo{
						Name:    damanager.GetKeyName(),
						Address: daAddress.Address,
					})
				}
			case consts.Walrus:
				// Initialize DAManager for Walrus
				damanager := datalayer.NewDAManager(
					consts.Walrus,
					home,
					localRollerConfig.KeyringBackend,
					localRollerConfig.NodeType,
				)

				// Retrieve DA account address
				daAddress, err := damanager.GetDAAccountAddress()
				if err != nil {
					pterm.Error.Println("failed to get Walrus account address: %w", err)
					return
				}
				// Append DA account address if available
				if daAddress != nil {
					addresses = append(addresses, keys.KeyInfo{
						Name:    damanager.GetKeyName(),
						Address: daAddress.Address,
					})
				}
			case consts.Solana:
				// Initialize DAManager for Solana
				damanager := datalayer.NewDAManager(
					consts.Solana,
					home,
					localRollerConfig.KeyringBackend,
					localRollerConfig.NodeType,
				)

				// Retrieve DA account address
				daAddress, err := damanager.GetDAAccountAddress()
				if err != nil {
					pterm.Error.Println("failed to get Solana account address: %w", err)
					return
				}
				// Append DA account address if available
				if daAddress != nil {
					addresses = append(addresses, keys.KeyInfo{
						Name:    damanager.GetKeyName(),
						Address: daAddress.Address,
					})
				}
			case consts.Ethereum:
				// Initialize DAManager for Ethereum
				damanager := datalayer.NewDAManager(
					consts.Ethereum,
					home,
					localRollerConfig.KeyringBackend,
					localRollerConfig.NodeType,
				)
				// Retrieve DA account address
				daAddress, err := damanager.GetDAAccountAddress()
				if err != nil {
					pterm.Error.Println("failed to get Ethereum account address: %w", err)
					return
				}

				// Append DA account address if available
				if daAddress != nil {
					addresses = append(addresses, keys.KeyInfo{
						Name:    damanager.GetKeyName(),
						Address: daAddress.Address,
					})
				}
			case consts.Kaspa:
				// Initialize DAManager for Kaspa
				damanager := datalayer.NewDAManager(
					consts.Kaspa,
					home,
					localRollerConfig.KeyringBackend,
					localRollerConfig.NodeType,
				)

				// Retrieve DA account address
				daAddress, err := damanager.GetDAAccountAddress()
				if err != nil {
					pterm.Error.Println("failed to get Kaspa account address: %w", err)
					return
				}

				// Append DA account address if available
				if daAddress != nil {
					addresses = append(addresses, keys.KeyInfo{
						Name:    damanager.GetKeyName(),
						Address: daAddress.Address,
					})
				}
			case consts.Mock:
			default:
				pterm.Error.Printf("unsupported DA backend: %s", rollappConfig.DA.Backend)
				return
			}

			damanager := datalayer.NewDAManager(
				rollappConfig.DA.Backend,
				rollappConfig.Home,
				rollappConfig.KeyringBackend,
				nodeType,
			)

			daHome := filepath.Join(
				damanager.GetRootDirectory(),
				consts.ConfigDirName.DALightNode,
			)

			isDaInitialized, err := filesystem.DirNotEmpty(daHome)
			if err != nil {
				return
			}

			var shouldOverwrite bool
			if isDaInitialized {
				pterm.Warning.Println("DA client is already initialized")
			}

			if !isDaInitialized || shouldOverwrite {
				mnemonic, err := damanager.InitializeLightNodeConfig()
				if err != nil {
					pterm.Error.Println("failed to initialize da light client: ", err)
					return
				}

				daWalletInfo, err := damanager.GetDAAccountAddress()
				if err != nil {
					pterm.Error.Println("failed to retrieve da wallet address: ", err)
					return
				}
				daWalletInfo.Mnemonic = mnemonic

				defer daWalletInfo.Print(keys.WithMnemonic(), keys.WithName())

				daSpinner, _ := pterm.DefaultSpinner.WithRemoveWhenDone(true).
					Start("initializing da light client")
				daSpinner.UpdateText("checking for state update ")
				cmd := exec.Command(
					consts.Executables.Dymension,
					"q",
					"rollapp",
					"state",
					rollappConfig.RollappID,
					"--index",
					"1",
					"--node",
					rollappConfig.HubData.RpcUrl,
					"--chain-id", rollappConfig.HubData.ID,
				)

				out, err := bash.ExecCommandWithStdout(cmd)
				if err != nil {
					if strings.Contains(out.String(), "key not found") {
						pterm.Info.Printf(
							"no state found for %s, da light client will be initialized with latest height",
							rollappConfig.RollappID,
						)

						height, blockIdHash, err := celestia.GetLatestBlock(localRollerConfig)
						if err != nil {
							return
						}

						heightInt, err := strconv.Atoi(height)
						if err != nil {
							pterm.Error.Println("failed to convert height to int: ", err)
							return
						}

						celestiaConfigFilePath := filepath.Join(
							home,
							consts.ConfigDirName.DALightNode,
							"config.toml",
						)

						pterm.Info.Printf("updating %s \n", celestiaConfigFilePath)
						err = celestialightclient.UpdateConfig(
							celestiaConfigFilePath,
							blockIdHash,
							heightInt,
						)
						if err != nil {
							pterm.Error.Println("failed to update celestia config: ", err)
							return
						}
					} else {
						pterm.Error.Println("failed to retrieve rollapp state update: ", err)
						return
					}
					// nolint:errcheck,gosec
					daSpinner.Stop()
				} else {
					daSpinner.UpdateText("state update found, extracting da height")
					// nolint:errcheck,gosec
					daSpinner.Stop()

					var result celestia.RollappStateResponse
					if err := yaml.Unmarshal(out.Bytes(), &result); err != nil {
						pterm.Error.Println("failed to unmarshal result: ", err)
						return
					}

					h, err := celestia.ExtractHeightfromDAPath(result.StateInfo.DAPath)
					if err != nil {
						pterm.Error.Println("failed to extract height: ", err)
						return
					}

					height, hash, err := celestia.GetBlockByHeight(h, localRollerConfig)
					if err != nil {
						pterm.Error.Println("failed to retrieve block: ", err)
						return
					}

					heightInt, err := strconv.Atoi(height)
					if err != nil {
						pterm.Error.Println("failed to convert height to int: ", err)
						return
					}

					celestiaConfigFilePath := filepath.Join(
						home,
						consts.ConfigDirName.DALightNode,
						"config.toml",
					)

					pterm.Info.Printf(
						"the first %s state update has DA height of %s with hash %s\n",
						rollappConfig.RollappID,
						height,
						hash,
					)
					pterm.Info.Printf("updating %s \n", celestiaConfigFilePath)
					err = celestialightclient.UpdateConfig(celestiaConfigFilePath, hash, heightInt)
					if err != nil {
						pterm.Error.Println("failed to update celestia config: ", err)
						return
					}
				}
			}

			dymintConfigPath := sequencer.GetDymintFilePath(home)
			appConfigPath := sequencer.GetAppConfigFilePath(home)

			switch nodeType {
			case "sequencer":
				pterm.Info.Println("checking DA account balance")
				insufficientBalances, err := damanager.CheckDABalance()
				if err != nil {
					pterm.Error.Println("failed to check balance", err)
				}

				err = tomlconfig.UpdateFieldInFile(
					dymintConfigPath,
					"p2p_advertising_enabled",
					"false",
				)
				if err != nil {
					pterm.Error.Println("failed to update `p2p_advertising_enabled`")
					return
				}

				err = keys.PrintInsufficientBalancesIfAny(insufficientBalances)
				if err != nil {
					pterm.Error.Println("failed to check insufficient balances: ", err)
					return
				}

			case "fullnode":

				vtu := map[string]any{
					"p2p_advertising_enabled": "true",
				}
				err := tomlconfig.UpdateFieldsInFile(dymintConfigPath, vtu)
				if err != nil {
					pterm.Error.Println("failed to update dymint config", err)
					return
				}

				var fullNodeType string
				if slices.Contains(fullNodeTypes, fullNodeTypeFromFlag) {
					fullNodeType = fullNodeTypeFromFlag
				} else {
					fullNodeType, _ = pterm.DefaultInteractiveSelect.
						WithDefaultText("select the environment you want to initialize for").
						WithOptions(fullNodeTypes).
						Show()
				}

				var fnVtu map[string]any

				switch fullNodeType {
				case "rpc":
					fnVtu = map[string]any{
						"pruning":             "custom",
						"pruning-keep-recent": "362880",
						"pruning-interval":    "100",
						"min-retain-blocks":   "362880",
					}
				case "archive":
					fnVtu = map[string]any{
						"pruning":           "nothing",
						"min-retain-blocks": "0",
					}
				}

				err = tomlconfig.UpdateFieldsInFile(appConfigPath, fnVtu)
				if err != nil {
					pterm.Error.Println("failed to update app config", err)
					return
				}
			default:
				pterm.Error.Println("unsupported node type")
				return
			}

			pterm.Info.Println("updating dymint configuration")

			_ = tomlconfig.UpdateFieldInFile(
				dymintConfigPath,
				"da_layer",
				getDaLayer(home, raResponse, damanager.DaType),
			)

			_ = tomlconfig.UpdateFieldInFile(
				dymintConfigPath,
				"da_config",
				getDaConfig(damanager.DataLayer, nodeType, home, raResponse, rollappConfig),
			)

			_ = tomlconfig.UpdateFieldInFile(
				dymintConfigPath,
				"max_proof_time",
				"5s",
			)
			_ = tomlconfig.UpdateFieldInFile(
				dymintConfigPath,
				"settlement_gas_prices",
				"20000000000adym",
			)

			pterm.Info.Println("enabling block explorer endpoint")
			_ = tomlconfig.UpdateFieldInFile(
				filepath.Join(home, consts.ConfigDirName.Rollapp, "config", "be-json-rpc.toml"),
				"enable",
				"true",
			)

			keys.PrintAddressesWithTitle(addresses)

			pterm.Info.Println("initialization complete")

			defer func() {
				pterm.Info.Println("next steps:")
				pterm.Info.Printf(
					"run %s load the necessary systemd services\n",
					pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
						Sprintf("roller rollapp services load"),
				)
			}()
		},
	}

	cmd.Flags().String("node-type", "", "node type ( supported values: [sequencer, fullnode] )")
	cmd.Flags().String("full-node-type", "", "full node type ( supported values: [rpc, archive] )")
	cmd.Flags().
		Bool("use-default-rpc-endpoint", false, "uses the default dymension hub rpc endpoint")

	return cmd
}

func SupportedGasDenoms(
	raCfg roller.RollappConfig,
) (map[string]dymensionseqtypes.DenomMetadata, error) {
	raResponse, err := rollapp.GetMetadataFromChain(
		raCfg.RollappID,
		raCfg.HubData,
	)
	if err != nil {
		pterm.Error.Println("failed to retrieve rollapp information from hub: ", err)
		return nil, err
	}
	sd := map[string]dymensionseqtypes.DenomMetadata{
		"ibc/FECACB927EB3102CCCB240FFB3B6FCCEEB8D944C6FEA8DFF079650FEFF59781D": {
			Display:  "DYM",
			Base:     "adym",
			Exponent: 18,
		},

		// "ibc/B3504E092456BA618CC28AC671A71FB08C6CA0FD0BE7C8A5B5A3E2DD933CC9E4": {
		// 	Display:  "usdc",
		// 	Base:     "ibc/B3504E092456BA618CC28AC671A71FB08C6CA0FD0BE7C8A5B5A3E2DD933CC9E4",
		// 	Exponent: 6,
		// },
	}

	if raResponse.Rollapp.GenesisInfo.NativeDenom != nil {
		sd[raResponse.Rollapp.GenesisInfo.NativeDenom.Base] = dymensionseqtypes.DenomMetadata{
			Display:  raResponse.Rollapp.GenesisInfo.NativeDenom.Display,
			Base:     raResponse.Rollapp.GenesisInfo.NativeDenom.Base,
			Exponent: raResponse.Rollapp.GenesisInfo.NativeDenom.Exponent,
		}
	}

	return sd, nil
}

func populateSequencerMetadata(raCfg roller.RollappConfig) error {
	cd := dymensionseqtypes.ContactDetails{
		Website:  "",
		Telegram: "",
		X:        "",
	}

	var dgpAmount string
	var ok bool

	as, err := genesis.GetAppStateFromGenesisFile(raCfg.Home)
	if err != nil {
		return err
	}

	if len(as.RollappParams.Params.MinGasPrices) == 0 {
		return errors.New("rollappparams should contain at least one gas token")
	}

	var denom string
	if len(as.RollappParams.Params.MinGasPrices) == 1 {
		dgpAmount = as.RollappParams.Params.MinGasPrices[0].String()
		denom = as.RollappParams.Params.MinGasPrices[0].Denom
	} else {
		pterm.Info.Println("more then 1 gas token option found")
		var options []string
		for _, token := range as.RollappParams.Params.MinGasPrices {
			options = append(options, token.Denom)
		}

		denom, _ := pterm.DefaultInteractiveSelect.WithOptions(options).WithDefaultText("select the token to use for the gas denom").Show()
		selectedIndex := slices.IndexFunc(as.RollappParams.Params.MinGasPrices, func(t cosmossdktypes.DecCoin) bool {
			return t.Denom == denom
		})
		dgpAmount = as.RollappParams.Params.MinGasPrices[selectedIndex].String()
	}

	sgt, err := SupportedGasDenoms(raCfg)
	if err != nil {
		return err
	}

	if _, ok = sgt[denom]; !ok {
		return errors.New("unsupported gas denom")
	}

	fd := sgt[denom]
	fd.Display = strings.ToUpper(fd.Display)

	// TODO: add support for other denoms
	var sm dymensionseqtypes.SequencerMetadata
	var defaultSnapshots []*dymensionseqtypes.SnapshotInfo

	if fd.Base != "adym" {
		sm = dymensionseqtypes.SequencerMetadata{
			Moniker:        "",
			Details:        "",
			P2PSeeds:       []string{},
			Rpcs:           []string{},
			EvmRpcs:        []string{},
			RestApiUrls:    []string{},
			ExplorerUrl:    "",
			GenesisUrls:    []string{},
			ContactDetails: &cd,
			ExtraData:      []byte{},
			Snapshots:      defaultSnapshots,
			GasPrice:       dgpAmount,
			FeeDenom:       nil,
		}
	} else {
		sm = dymensionseqtypes.SequencerMetadata{
			Moniker:        "",
			Details:        "",
			P2PSeeds:       []string{},
			Rpcs:           []string{},
			EvmRpcs:        []string{},
			RestApiUrls:    []string{},
			ExplorerUrl:    "",
			GenesisUrls:    []string{},
			ContactDetails: &cd,
			ExtraData:      []byte{},
			Snapshots:      defaultSnapshots,
			GasPrice:       dgpAmount,
			FeeDenom:       &fd,
		}
	}

	path := filepath.Join(
		raCfg.Home,
		consts.ConfigDirName.Rollapp,
		"init",
		"sequencer-metadata.json",
	)
	pterm.DefaultSection.WithIndentCharacter("🔔").
		Println("The following values are mandatory for sequencer creation")

	var rpc string
	var rest string
	var evmRpc string

	for {
		// Prompt the user for the RPC URL
		rpc, _ = pterm.DefaultInteractiveTextInput.WithDefaultText(
			"rollapp rpc endpoint that you will provide (example: https://rpc.rollapp.dym.xyz:443)",
		).Show()
		if !strings.HasPrefix(rpc, "http://") && !strings.HasPrefix(rpc, "https://") {
			rpc = "https://" + rpc
		}

		isValid := config.IsValidURL(rpc)

		// Validate the URL
		if !isValid {
			pterm.Error.Println("Invalid URL. Please try again.")
		} else {
			// Valid URL, break out of the loop
			break
		}
	}

	for {
		// Prompt the user for the RPC URL
		rest, _ = pterm.DefaultInteractiveTextInput.WithDefaultText(
			"rest endpoint that you will provide (example: https://api.rollapp.dym.xyz:443)",
		).Show()
		if !strings.HasPrefix(rest, "http://") && !strings.HasPrefix(rest, "https://") {
			rest = "https://" + rest
		}

		isValid := config.IsValidURL(rest)

		// Validate the URL
		if !isValid {
			pterm.Error.Println("Invalid URL. Please try again.")
		} else {
			// Valid URL, break out of the loop
			break
		}
	}

	if raCfg.RollappVMType == consts.EVM_ROLLAPP {
		for {
			// Prompt the user for the RPC URL
			evmRpc, _ = pterm.DefaultInteractiveTextInput.WithDefaultText(
				"evm rpc endpoint that you will provide (example: https://json-rpc.rollapp.dym.xyz:443)",
			).Show()
			if !strings.HasPrefix(evmRpc, "http://") && !strings.HasPrefix(evmRpc, "https://") {
				evmRpc = "https://" + evmRpc
			}

			isValid := config.IsValidURL(evmRpc)

			// Validate the URL
			if !isValid {
				pterm.Error.Println("Invalid URL. Please try again.")
			} else {
				// Valid URL, break out of the loop
				break
			}
		}
		sm.EvmRpcs = append(sm.EvmRpcs, evmRpc)
	}

	sm.Rpcs = append(sm.Rpcs, rpc)
	sm.RestApiUrls = append(sm.RestApiUrls, rest)

	shouldFillOptionalFields, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(
		"Would you also like to fill optional metadata for your sequencer?",
	).Show()

	if shouldFillOptionalFields {
		displayName, _ := pterm.DefaultInteractiveTextInput.WithDefaultText(
			"provide a display name for your sequencer",
		).Show()
		x, _ := pterm.DefaultInteractiveTextInput.WithDefaultText(
			"provide a link to your X",
		).Show()
		website, _ := pterm.DefaultInteractiveTextInput.WithDefaultText(
			"provide a link to your website",
		).Show()
		if !strings.HasPrefix(website, "http://") && !strings.HasPrefix(website, "https://") {
			website = "https://" + website
		}
		telegram, _ := pterm.DefaultInteractiveTextInput.WithDefaultText(
			"provide a link to your telegram",
		).Show()

		sm.ContactDetails.X = x
		sm.ContactDetails.Website = website
		sm.ContactDetails.Telegram = telegram
		sm.Moniker = displayName
	}

	err = WriteStructToJSONFile(&sm, path)
	if err != nil {
		return err
	}
	return nil
}

func WriteStructToJSONFile(data *dymensionseqtypes.SequencerMetadata, filePath string) error {
	// Marshal the struct into JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	// Create the directory path if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("error creating directories: %w", err)
	}

	// Write the JSON data to the file
	if err := os.WriteFile(filePath, jsonData, 0o644); err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	return nil
}

func displayRegularDenom(coin cosmossdktypes.Coin, decimals int) string {
	decCoin := cosmossdktypes.NewDecCoinFromCoin(coin)

	// Create a divisor (10^18)
	divisor := cosmossdktypes.NewDec(10).Power(uint64(decimals))

	// Divide the amount
	amount := decCoin.Amount.Quo(divisor)

	// Format the amount with 6 decimal places (or adjust as needed)
	formattedAmount := amount.String()
	if strings.Contains(formattedAmount, ".") {
		parts := strings.Split(formattedAmount, ".")
		if len(parts[1]) > 18 {
			formattedAmount = fmt.Sprintf("%s.%s", parts[0], parts[1][:18])
		}
	}

	return formattedAmount
}

func getDaLayer(home string, raResponse *rollapp.ShowRollappResponse, daType consts.DAType) any {
	drsVersion, err := genesis.GetDrsVersionFromGenesis(home, raResponse)
	if err != nil {
		pterm.Error.Println("failed to get drs version from genesis: ", err)
		return nil
	}

	if rollapp.IsDaConfigNewFormat(drsVersion, strings.ToLower(raResponse.Rollapp.VmType)) {
		return []string{string(daType)}
	} else {
		return daType
	}
}

func getDaConfig(
	dataLayer datalayer.DataLayer,
	nodeType string,
	home string,
	raResponse *rollapp.ShowRollappResponse,
	rollappConfig *roller.RollappConfig,
) any {
	daConfig := dataLayer.GetSequencerDAConfig(nodeType)

	drsVersion, err := genesis.GetDrsVersionFromGenesis(home, raResponse)
	if err != nil {
		pterm.Error.Println("failed to get drs version from genesis: ", err)
		return nil
	}

	if rollapp.IsDaConfigNewFormat(drsVersion, strings.ToLower(raResponse.Rollapp.VmType)) {
		return []string{daConfig}
	} else {
		return daConfig
	}
}
