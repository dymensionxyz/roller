package setup

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	cosmossdkmath "cosmossdk.io/math"
	cosmossdktypes "github.com/cosmos/cosmos-sdk/types"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	dymensionseqtypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	"github.com/dymensionxyz/roller/data_layer/celestia"
	"github.com/dymensionxyz/roller/data_layer/celestia/lightclient"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/rollapp"
	"github.com/dymensionxyz/roller/utils/roller"
	"github.com/dymensionxyz/roller/utils/sequencer"
)

// TODO: Test sequencing on 35-C and update the price
var OneDaySequencePrice = big.NewInt(1)

var (
	RollappDirPath string
	LogPath        string
)

// nolint:gocyclo
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Setup a RollApp node.",
		Long:  ``,
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
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

			rollappConfig, err := rollapp.PopulateRollerConfigWithRaMetadataFromChain(
				home,
				localRollerConfig.RollappID,
				localRollerConfig.HubData,
			)
			errorhandling.PrettifyErrorIfExists(err)

			if rollappConfig.HubData.ID == "mock" {
				pterm.Error.Println("setup is not required for mock backend")
				pterm.Info.Printf(
					"run %s instead to run the rollapp\n",
					pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
						Sprintf("roller rollapp start"),
				)
				return
			}

			raResponse, err := rollapp.GetMetadataFromChain(
				localRollerConfig.RollappID,
				localRollerConfig.HubData,
			)
			if err != nil {
				pterm.Error.Println("failed to fetch rollapp information from hub: ", err)
				return
			}

			if raResponse.Rollapp.PreLaunchTime != "" {
				timeLayout := time.RFC3339Nano
				expectedLaunchTime, err := time.Parse(timeLayout, raResponse.Rollapp.PreLaunchTime)
				if err != nil {
					pterm.Error.Println("failed to parse launch time", err)
					return
				}

				if expectedLaunchTime.After(time.Now()) {
					pterm.Error.Printf(
						`Nodes can be set up only after the minimum IRO duration has passed
Current time: %v
RollApp's IRO time: %v`,
						time.Now().UTC().Format(timeLayout),
						expectedLaunchTime.Format(timeLayout),
					)

					return
				}
			} else {
				pterm.Info.Printf("no IRO set up for %s\n", raResponse.Rollapp.RollappId)
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
				return
			}

			options := []string{"sequencer", "fullnode"}
			nodeType, _ := pterm.DefaultInteractiveSelect.
				WithDefaultText("select the node type you want to run").
				WithOptions(options).
				Show()

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
					minBond, _ := sequencer.GetMinSequencerBondInBaseDenom(rollappConfig.HubData)
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

						if err != nil {
							pterm.Error.Println("failed to convert desired bond amount to base denom: ", err)
							return
						}
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

					pterm.Info.Printf(
						"current balance: %s\nnecessary balance: %s\n",
						balance.String(),
						fmt.Sprintf("%s%s", necessaryBalance.String(), consts.Denoms.Hub),
					)

					// check whether balance is bigger or equal to the necessaryBalance
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

					pterm.Info.Printf(
						"current balance: %s\nnecessary balance: %s\n",
						balance.String(),
						fmt.Sprintf("%s%s", necessaryBalance.String(), consts.Denoms.Hub),
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

				// approve the data directory deletion before downloading the snapshot,
				dataDir := filepath.Join(RollappDirPath, "data")
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

						err = filesystem.ExtractTarGz(archivePath, filepath.Join(RollappDirPath))
						if err != nil {
							pterm.Error.Println("failed to extract snapshot: ", err)
							return
						}
					}
				}
			}

			// DA
			damanager := datalayer.NewDAManager(
				rollappConfig.DA.Backend,
				rollappConfig.Home,
				rollappConfig.KeyringBackend,
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
				daWalletInfo.Print(keys.WithMnemonic(), keys.WithName())

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
						err = lightclient.UpdateConfig(
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

					var result lightclient.RollappStateResponse
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
					err = lightclient.UpdateConfig(celestiaConfigFilePath, hash, heightInt)
					if err != nil {
						pterm.Error.Println("failed to update celestia config: ", err)
						return
					}
				}
			}

			var daConfig string
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

				// TODO: daconfig should be a struct
				daConfig = damanager.DataLayer.GetSequencerDAConfig(
					consts.NodeType.Sequencer,
				)

			case "fullnode":
				daConfig = damanager.DataLayer.GetSequencerDAConfig(
					consts.NodeType.FullNode,
				)

				vtu := map[string]any{
					"p2p_advertising_enabled": "true",
				}
				err := tomlconfig.UpdateFieldsInFile(dymintConfigPath, vtu)
				if err != nil {
					pterm.Error.Println("failed to update dymint config", err)
					return
				}

				fullNodeTypes := []string{"rpc", "archive"}
				fullNodeType, _ := pterm.DefaultInteractiveSelect.
					WithDefaultText("select the environment you want to initialize for").
					WithOptions(fullNodeTypes).
					Show()
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
						"pruning": "nothing",
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

			daNamespace := damanager.DataLayer.GetNamespaceID()
			if daNamespace == "" {
				pterm.Error.Println("failed to retrieve da namespace id")
				return
			}

			pterm.Info.Println("updating dymint configuration")
			_ = tomlconfig.UpdateFieldInFile(
				dymintConfigPath,
				"da_layer",
				string(rollappConfig.DA.Backend),
			)
			_ = tomlconfig.UpdateFieldInFile(
				dymintConfigPath,
				"namespace_id",
				daNamespace,
			)
			_ = tomlconfig.UpdateFieldInFile(
				dymintConfigPath,
				"da_config",
				daConfig,
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

	return cmd
}

func populateSequencerMetadata(raCfg roller.RollappConfig) error {
	cd := dymensionseqtypes.ContactDetails{
		Website:  "",
		Telegram: "",
		X:        "",
	}

	var defaultGasPrice cosmossdktypes.Int
	var ok bool

	if raCfg.HubData.GasPrice != "" {
		defaultGasPrice, ok = github_com_cosmos_cosmos_sdk_types.NewIntFromString(
			raCfg.HubData.GasPrice,
		)
	} else {
		defaultGasPrice = cosmossdktypes.NewInt(2000000000)
		ok = true
	}
	if !ok {
		return errors.New("failed to parse gas price")
	}

	var defaultSnapshots []*dymensionseqtypes.SnapshotInfo
	sm := dymensionseqtypes.SequencerMetadata{
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
		GasPrice:       &defaultGasPrice,
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
			"dymint rpc endpoint that you will provide (example: rpc.rollapp.dym.xyz)",
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
			"rest endpoint that you will provide (example: api.rollapp.dym.xyz)",
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

	for {
		// Prompt the user for the RPC URL
		evmRpc, _ = pterm.DefaultInteractiveTextInput.WithDefaultText(
			"evm rpc endpoint that you will provide (example: json-rpc.rollapp.dym.xyz)",
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

	sm.Rpcs = append(sm.Rpcs, rpc)
	sm.RestApiUrls = append(sm.RestApiUrls, rest)
	sm.EvmRpcs = append(sm.EvmRpcs, evmRpc)

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

	err := WriteStructToJSONFile(&sm, path)
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
