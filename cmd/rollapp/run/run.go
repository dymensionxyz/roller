package run

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	cosmossdkmath "cosmossdk.io/math"
	cosmossdktypes "github.com/cosmos/cosmos-sdk/types"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	dymensionseqtypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	initrollapp "github.com/dymensionxyz/roller/cmd/rollapp/init"
	"github.com/dymensionxyz/roller/cmd/utils"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	"github.com/dymensionxyz/roller/sequencer"
	globalutils "github.com/dymensionxyz/roller/utils"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	sequencerutils "github.com/dymensionxyz/roller/utils/sequencer"
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
		Use:   "run [rollapp-id]",
		Short: "Run the RollApp nodes",
		Long:  ``,
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := initconfig.AddFlags(cmd)
			if err != nil {
				pterm.Error.Println("failed to add flags")
				return
			}

			home, err := globalutils.ExpandHomePath(cmd.Flag(utils.FlagNames.Home).Value.String())
			if err != nil {
				pterm.Error.Println("failed to expand home directory")
				return
			}

			rollerData, err := tomlconfig.LoadRollerConfig(home)
			if err != nil {
				pterm.Error.Println("failed to load roller config file", err)
				return
			}

			hd, err := tomlconfig.LoadHubData(home)
			if err != nil {
				pterm.Error.Println("failed to load hub data from roller.toml")
			}

			rollappConfig, err := tomlconfig.LoadRollappMetadataFromChain(
				home,
				rollerData.RollappID,
				&hd,
			)
			errorhandling.PrettifyErrorIfExists(err)

			seq := sequencer.GetInstance(*rollappConfig)
			startRollappCmd := seq.GetStartCmd()

			LogPath = filepath.Join(
				rollappConfig.Home,
				consts.ConfigDirName.Rollapp,
				"rollapputils.log",
			)
			RollappDirPath = filepath.Join(rollappConfig.Home, consts.ConfigDirName.Rollapp)

			if rollappConfig.HubData.ID == "mock" {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				go bash.RunCmdAsync(
					ctx, startRollappCmd, func() {
						printOutput(*rollappConfig, startRollappCmd)
						err := createPidFile(RollappDirPath, startRollappCmd)
						if err != nil {
							pterm.Warning.Println("failed to create pid file")
						}
					}, parseError,
					utils.WithLogging(utils.GetSequencerLogPath(*rollappConfig)),
				)
				select {}
			}

			options := []string{"sequencer", "fullnode"}
			nodeType, _ := pterm.DefaultInteractiveSelect.
				WithDefaultText("select the node type you want to run").
				WithOptions(options).
				Show()

			switch nodeType {
			case "sequencer":
				pterm.Info.Println("getting the existing sequencer address ")
				hubSeqKC := utils.KeyConfig{
					Dir:         filepath.Join(rollappConfig.Home, consts.ConfigDirName.HubKeys),
					ID:          consts.KeysIds.HubSequencer,
					ChainBinary: consts.Executables.Dymension,
					Type:        consts.SDK_ROLLAPP,
				}
				seqAddrInfo, err := utils.GetAddressInfoBinary(hubSeqKC, hubSeqKC.ChainBinary)
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

				seq, err := sequencerutils.GetRegisteredSequencers(rollappConfig.RollappID, hd)
				if err != nil {
					pterm.Error.Println("failed to retrieve registered sequencers: ", err)
				}

				isSequencerRegistered := sequencerutils.IsRegisteredAsSequencer(
					seq.Sequencers,
					seqAddrInfo.Address,
				)

				if !isSequencerRegistered {
					minBond, _ := sequencerutils.GetMinSequencerBond()
					var bondAmount cosmossdktypes.Coin
					bondAmount.Denom = consts.Denoms.Hub

					var desiredBond cosmossdktypes.Coin
					desiredBondAmount, _ := pterm.DefaultInteractiveTextInput.WithDefaultText(
						fmt.Sprintf(
							"what is your desired bond amount? ( min: %s ) press enter to proceed with %s",
							minBond.String(),
							minBond.String(),
						),
					).WithDefaultValue(minBond.Amount.String()).Show()

					if strings.TrimSpace(desiredBondAmount) == "" {
						desiredBond = *minBond
					} else {
						desiredBondAmountInt, ok := cosmossdkmath.NewIntFromString(desiredBondAmount)
						if !ok {
							pterm.Error.Printf("failed to convert %s to int\n", desiredBondAmount)
							return
						}

						desiredBond.Denom = consts.Denoms.Hub
						desiredBond.Amount = desiredBondAmountInt
					}

					pterm.Info.Println("getting the existing sequencer address balance")
					balance, err := utils.QueryBalance(
						utils.ChainQueryConfig{
							Denom:  consts.Denoms.Hub,
							RPC:    rollappConfig.HubData.RPC_URL,
							Binary: consts.Executables.Dymension,
						}, seqAddrInfo.Address,
					)
					if err != nil {
						pterm.Error.Println("failed to get address balance: ", err)
						return
					}

					// TODO: use NotFundedAddressData instead
					var necessaryBalance big.Int
					necessaryBalance.Add(
						desiredBond.Amount.BigInt(),
						cosmossdkmath.NewInt(consts.DefaultFee).BigInt(),
					)

					pterm.Info.Printf(
						"current balance: %s\nnecessary balance: %s\n",
						balance.Amount.String(),
						necessaryBalance.String(),
					)

					// check whether balance is bigger or equal to the necessaryBalance
					isAddrFunded := balance.Amount.Cmp(&necessaryBalance) == 1 ||
						balance.Amount.Cmp(
							&necessaryBalance,
						) == 0

					if !isAddrFunded {
						pterm.DefaultSection.WithIndentCharacter("🔔").
							Println("Please fund the addresses below to register and run the sequencer.")
						seqAddrInfo.Print(utils.WithName())
						proceed, _ := pterm.DefaultInteractiveConfirm.WithDefaultValue(true).
							WithDefaultText(
								"press enter when funded",
							).Show()

						if !proceed {
							return
						}
					}

					// isInitialSequencer, err := rollapputils.IsInitialSequencer(
					// 	seqAddrInfo.Address,
					// 	rollappConfig.RollappID,
					// )
					// if err != nil {
					// 	pterm.Error.Printf(
					// 		"failed to check whether %s is the initial sequencer\n",
					// 		seqAddrInfo.Address,
					// 	)
					// }

					// if isInitialSequencer {
					// 	pterm.Info.Printf(
					// 		"the %s ( %s ) address matches the initial sequencer address of the %s\n",
					// 		seqAddrInfo.Name,
					// 		seqAddrInfo.Address,
					// 		rollappConfig.RollappID,
					// 	)
					// 	pterm.Info.Printf(
					// 		"initial sequencer address is not registered for %s\n",
					// 		rollappConfig.RollappID,
					// 	)

					err = populateSequencerMetadata(*rollappConfig)
					if err != nil {
						pterm.Error.Println("failed to populate sequencer metadata: ", err)
						return
					}

					err = sequencerutils.Register(*rollappConfig)
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
					// } else {
					// 	pterm.Info.Printf(
					// 		"%s ( %s ) is not the initial sequencer address\n",
					// 		seqAddrInfo.Name,
					// 		seqAddrInfo.Address,
					// 	)
					//
					// 	pterm.Info.Printf(
					// 		"checking whether the initial sequencer is already registered for %s\n",
					// 		rollappConfig.RollappID,
					// 	)
					// 	initialSeqAddr, err := rollapputils.GetInitialSequencerAddress(rollappConfig.RollappID)
					// 	if err != nil {
					// 		pterm.Error.Println("failed to retrieve initial sequencer address: ", err)
					// 		return
					// 	}
					//
					// 	isInitialSequencerRegistered := sequencerutils.IsRegisteredAsSequencer(
					// 		seq.Sequencers,
					// 		initialSeqAddr,
					// 	)
					//
					// 	if !isInitialSequencerRegistered {
					// 		pterm.Warning.Println("additional sequencers can only be added after the initial sequencer is registered")
					// 		pterm.Info.Println("exiting")
					// 		return
					// 	}
					//
					// 	pterm.Info.Println(
					// 		"initial sequencer is already registered, proceeding with creation of your sequencer",
					// 	)
					//
					// 	err = populateSequencerMetadata(rollappConfig)
					// 	if err != nil {
					// 		pterm.Error.Println("failed to populate sequencer metadata: ", err)
					// 		return
					// 	}
					// 	err = sequencerutils.Register(rollappConfig)
					// 	if err != nil {
					// 		pterm.Error.Println("failed to register sequencer: ", err)
					// 		return
					// 	}
					// 	pterm.Info.Printf(
					// 		"%s ( %s ) is registered as a sequencer for %s\n",
					// 		seqAddrInfo.Name,
					// 		seqAddrInfo.Address,
					// 		rollappConfig.RollappID,
					// 	)
					// }
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
				si, err := sequencerutils.GetLatestSnapshot(rollappConfig.RollappID, hd)
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
						"found a snapshot for height %s\nchecksum: %s\nurl: %s",
						si.Height,
						si.Checksum,
						si.SnapshotUrl,
					)
				}

				// look for p2p bootstrap nodes, if there are no nodes available, the rollapp
				// defaults to syncing only from the DA
				peers, err := sequencerutils.GetAllP2pPeers(rollappConfig.RollappID, hd)
				if err != nil {
					pterm.Error.Println("failed to retrieve p2p peers ")
				}

				if len(peers) == 0 {
					pterm.Warning.Println(
						"none of the sequencers provide p2p seed nodes this node will sync only from DA",
					)
				}

				// approve the data directory deletion before downloading the snapshot,
				dataDir := filepath.Join(RollappDirPath, "data")
				if fi, err := os.Stat(dataDir); err == nil && fi.IsDir() {
					dataDirNotEmpty, err := globalutils.DirNotEmpty(dataDir)
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
						defer os.RemoveAll(tmpDir)
						spinner, _ := pterm.DefaultSpinner.Start("downloading file...")
						downloadedFileHash, err := globalutils.DownloadAndSaveArchive(
							si.SnapshotUrl,
							tmpDir,
						)
						if err != nil {
							spinner.Fail(fmt.Sprintf("error downloading file: %v", err))
							os.Exit(1)
						}
						spinner.Success("file downloaded successfully")

						// compare the checksum
						if downloadedFileHash != si.Checksum {
							pterm.Error.Println()
						}

						err = globalutils.ExtractTarGz(tmpDir, filepath.Join(RollappDirPath))
						if err != nil {
							pterm.Error.Println("failed to extract snapshot: ", err)
							return
						}
					}
				}
			}

			// DA
			oh := initconfig.NewOutputHandler(false)
			damanager := datalayer.NewDAManager(rollappConfig.DA, rollappConfig.Home)
			daHome := filepath.Join(
				damanager.GetRootDirectory(),
				consts.ConfigDirName.DALightNode,
			)

			isDaInitialized, err := globalutils.DirNotEmpty(daHome)
			if err != nil {
				return
			}

			var shouldOverwrite bool
			if isDaInitialized {
				pterm.Info.Println("DA client is already initialized")
				oh.StopSpinner()
				shouldOverwrite, err = oh.PromptOverwriteConfig(daHome)
				if err != nil {
					return
				}
			}

			if shouldOverwrite {
				pterm.Info.Println("overriding the existing da configuration")
				err := os.RemoveAll(daHome)
				if err != nil {
					pterm.Error.Printf("failed to remove %s: %v\n", daHome, err)
					return
				}

				err = os.MkdirAll(daHome, 0o755)
				if err != nil {
					pterm.Error.Printf("failed to create %s: %v\n", daHome, err)
					return
				}
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
				daWalletInfo.Print(utils.WithMnemonic(), utils.WithName())

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
					hd.RPC_URL,
				)

				out, err := bash.ExecCommandWithStdout(cmd)
				if err != nil {
					if strings.Contains(out.String(), "key not found") {
						pterm.Info.Printf(
							"no state found for %s, da light client will be initialized with latest height",
							rollappConfig.RollappID,
						)

						height, blockIdHash, err := initrollapp.GetLatestDABlock()
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
						err = initrollapp.UpdateCelestiaConfig(
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

					var result initrollapp.Result
					if err := yaml.Unmarshal(out.Bytes(), &result); err != nil {
						pterm.Error.Println("failed to unmarshal result: ", err)
						return
					}

					h, err := initrollapp.ExtractHeightfromDAPath(result.StateInfo.DAPath)
					if err != nil {
						pterm.Error.Println("failed to extract height: ", err)
						return
					}

					height, hash, err := initrollapp.GetDABlockByHeight(h)
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
					err = initrollapp.UpdateCelestiaConfig(celestiaConfigFilePath, hash, heightInt)
					if err != nil {
						pterm.Error.Println("failed to update celestia config: ", err)
						return
					}
				}
			}

			var daConfig string

			switch nodeType {
			case "sequencer":
				pterm.Info.Println("checking DA account balance")
				insufficientBalances, err := damanager.CheckDABalance()
				if err != nil {
					pterm.Error.Println("failed to check balance", err)
				}

				utils.PrintInsufficientBalancesIfAny(insufficientBalances)

				// TODO: daconfig should be a struct
				daConfig = damanager.DataLayer.GetSequencerDAConfig(
					consts.NodeType.Sequencer,
				)

			case "fullnode":
				daConfig = damanager.DataLayer.GetSequencerDAConfig(
					consts.NodeType.FullNode,
				)
			default:
				pterm.Error.Println("unsupported node type")
				return

			}

			dymintConfigPath := sequencer.GetDymintFilePath(home)
			daNamespace := damanager.DataLayer.GetNamespaceID()
			if daNamespace == "" {
				pterm.Error.Println("failed to retrieve da namespace id")
				return
			}

			pterm.Info.Println("updating dymint configuration")
			_ = globalutils.UpdateFieldInToml(
				dymintConfigPath,
				"da_layer",
				string(rollappConfig.DA),
			)
			_ = globalutils.UpdateFieldInToml(
				dymintConfigPath,
				"namespace_id",
				daNamespace,
			)
			_ = globalutils.UpdateFieldInToml(
				dymintConfigPath,
				"da_config",
				daConfig,
			)

			pterm.Info.Println("initialization complete")
			pterm.Info.Println("next steps:")
			pterm.Info.Printf(
				"run %s load the necessary systemd services\n",
				pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
					Sprintf("roller rollapp services load"),
			)
		},
	}

	return cmd
}

func printOutput(rlpCfg config.RollappConfig, cmd *exec.Cmd) {
	seq := sequencer.GetInstance(rlpCfg)
	pterm.DefaultSection.WithIndentCharacter("💈 ").
		Println("The Rollapp sequencer is running on your local machine!")
	fmt.Println("💈 Endpoints:")

	fmt.Printf("EVM RPC: http://127.0.0.1:%v\n", seq.JsonRPCPort)
	fmt.Printf("Node RPC: http://127.0.0.1:%v\n", seq.RPCPort)
	fmt.Printf("Rest API: http://127.0.0.1:%v\n", seq.APIPort)

	fmt.Println("💈 Log file path: ", LogPath)
	fmt.Println("💈 Rollapp root dir: ", RollappDirPath)
	fmt.Printf(
		"💈 PID: %d (saved in %s)\n",
		cmd.Process.Pid,
		filepath.Join(rlpCfg.Home, "rollapputils.pid"),
	)
}

func createPidFile(path string, cmd *exec.Cmd) error {
	pidPath := filepath.Join(path, "rollapputils.pid")
	file, err := os.Create(pidPath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return err
	}
	// nolint errcheck
	defer file.Close()

	pid := cmd.Process.Pid
	_, err = file.WriteString(fmt.Sprintf("%d", pid))
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return err
	}

	return nil
}

func parseError(errMsg string) string {
	lines := strings.Split(errMsg, "\n")
	if len(lines) > 0 &&
		lines[0] == "Error: failed to initialize database: resource temporarily unavailable" {
		return "The Rollapp sequencer is already running on your local machine. Only one sequencer can run at any given time."
	}
	return errMsg
}

func populateSequencerMetadata(raCfg config.RollappConfig) error {
	cd := dymensionseqtypes.ContactDetails{
		Website:  "",
		Telegram: "",
		X:        "",
	}
	defaultGasPrice, ok := github_com_cosmos_cosmos_sdk_types.NewIntFromString(
		raCfg.HubData.GAS_PRICE,
	)
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

	// todo: clean up

	for {
		// Prompt the user for the RPC URL
		rpc, _ = pterm.DefaultInteractiveTextInput.WithDefaultText(
			"Enter a valid RPC endpoint (example: rpc.rollapp.dym.xyz)",
		).Show()
		if !strings.HasPrefix(rpc, "http://") && !strings.HasPrefix(rpc, "https://") {
			rpc = "https://" + rpc
		}

		isValid := isValidURL(rpc)

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
			"rest endpoint that you will provide (example: api.rollapp.dym.xyz",
		).Show()
		if !strings.HasPrefix(rest, "http://") && !strings.HasPrefix(rest, "https://") {
			rest = "https://" + rest
		}

		isValid := isValidURL(rest)

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
			"evm evmRpc endpoint that you will provide (example: json-rpc.rollapp.dym.xyz",
		).Show()
		if !strings.HasPrefix(evmRpc, "http://") && !strings.HasPrefix(evmRpc, "https://") {
			evmRpc = "https://" + evmRpc
		}

		isValid := isValidURL(evmRpc)

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

	_, _ = pterm.DefaultInteractiveConfirm.WithDefaultText(
		"Would you also like to fill optional metadata for your sequencer?",
	).Show()

	err := WriteStructToJSONFile(&sm, path)
	if err != nil {
		return err
	}
	return nil
}

func isValidURL(url string) bool {
	regex := `^(https?:\/\/)?([\da-z\.-]+)\.([a-z\.]{2,6})([\/\w \.-]*)*\/?$`
	re := regexp.MustCompile(regex)
	return re.MatchString(url)
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
	if err := ioutil.WriteFile(filePath, jsonData, 0o644); err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	return nil
}
