package run

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	cosmossdkmath "cosmossdk.io/math"
	cosmossdktypes "github.com/cosmos/cosmos-sdk/types"
	dymensionseqtypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	"github.com/dymensionxyz/roller/sequencer"
	globalutils "github.com/dymensionxyz/roller/utils"
	rollapputils "github.com/dymensionxyz/roller/utils/rollapp"
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
		Use:   "run",
		Short: "Initialize RollApp locally",
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

			rollappConfig, err := config.LoadRollerConfigFromTOML(home)
			utils.PrettifyErrorIfExists(err)

			seq := sequencer.GetInstance(rollappConfig)
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
				go utils.RunBashCmdAsync(
					ctx, startRollappCmd, func() {
						printOutput(rollappConfig, startRollappCmd)
						err := createPidFile(RollappDirPath, startRollappCmd)
						if err != nil {
							pterm.Warning.Println("failed to create pid file")
						}
					}, parseError,
					utils.WithLogging(utils.GetSequencerLogPath(rollappConfig)),
				)
				select {}
			}

			options := []string{"sequencer", "fullnode"}
			nodeType, _ := pterm.DefaultInteractiveSelect.
				WithDefaultText("select the settlement layer backend").
				WithOptions(options).
				Show()

			switch nodeType {
			case "sequencer":
				pterm.Info.Println("getting the existing sequencer address ")

				hubSeqKC := utils.KeyConfig{
					Dir:         filepath.Join(rollappConfig.Home, consts.ConfigDirName.HubKeys),
					ID:          consts.KeysIds.HubSequencer,
					ChainBinary: consts.Executables.Dymension,
					Type:        config.SDK_ROLLAPP,
				}
				seqAddrInfo, err := utils.GetAddressInfoBinary(hubSeqKC, hubSeqKC.ChainBinary)
				if err != nil {
					pterm.Error.Println("failed to get address info: ", err)
					return
				}
				seqAddrInfo.Address = strings.TrimSpace(seqAddrInfo.Address)

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

				var necessaryBalance big.Int
				necessaryBalance.Add(
					desiredBond.Amount.BigInt(),
					cosmossdkmath.NewInt(consts.DefaultFee).BigInt(),
				)

				isAddrFunded := balance.Amount.Cmp(&necessaryBalance) == 1

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

				isInitialSequencer, err := rollapputils.IsInitialSequencer(
					seqAddrInfo.Address,
					rollappConfig.RollappID,
				)
				if err != nil {
					pterm.Error.Printf(
						"failed to check whether %s is the initial sequencer\n",
						seqAddrInfo.Address,
					)
				}

				if isInitialSequencer {
					pterm.Info.Printf(
						"the %s ( %s ) address matches the initial sequencer address of the %s\n",
						seqAddrInfo.Name,
						seqAddrInfo.Address,
						rollappConfig.RollappID,
					)
					pterm.Info.Println(
						"checking whether sequencer is already registered",
						rollappConfig.RollappID,
					)

					seq, err := rollapputils.GetRegisteredSequencers(rollappConfig.RollappID)
					if err != nil {
						pterm.Error.Println("failed to retrieve registered sequencers: ", err)
					}

					isInitialSequencerRegistered := sequencerutils.IsRegisteredAsSequencer(
						seq.Sequencers,
						seqAddrInfo.Address,
					)
					if !isInitialSequencerRegistered {
						pterm.Info.Printf(
							"initial sequencer address is not registered for %s\n",
							rollappConfig.RollappID,
						)

						var sm dymensionseqtypes.SequencerMetadata

						seqMetadataPath := filepath.Join(
							rollappConfig.Home,
							consts.ConfigDirName.Rollapp,
							"init",
							"sequencer-metadata.json",
						)

						err = populateSequencerMetadata(sm, seqMetadataPath)
						if err != nil {
							pterm.Error.Println("failed to populate sequencer metadata: ", err)
							return
						}

						err = sequencerutils.Register(rollappConfig)
						if err != nil {
							pterm.Error.Println("failed to register sequencer: ", err)
							return
						}
					}
					pterm.Info.Printf(
						"%s ( %s ) is registered as a sequencer for %s\n",
						seqAddrInfo.Name,
						seqAddrInfo.Address,
						rollappConfig.RollappID,
					)
				} else {
					pterm.Info.Printf(
						"%s ( %s ) is not the initial sequencer address\n",
						seqAddrInfo.Name,
						seqAddrInfo.Address,
					)
				}

			case "fullnode":
				pterm.Info.Println("getting the fullnode address ")
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
					pterm.Error.Printf("failed to recuresively remove %s: %v\n", daHome, err)
					return
				}

				err = os.MkdirAll(daHome, 0o755)
				if err != nil {
					pterm.Error.Printf("failed to create %s: %v\n", daHome, err)
					return
				}
			}

			if !isDaInitialized || shouldOverwrite {
				if rollappConfig.DA == "celestia" {
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

					if nodeType == "sequencer" {
						pterm.DefaultSection.WithIndentCharacter("🔔").
							Println("Please fund the addresses below to register and run the sequencer.")
						daWalletInfo.Print(utils.WithMnemonic(), utils.WithName())

						proceed, _ := pterm.DefaultInteractiveConfirm.WithDefaultValue(true).
							WithDefaultText(
								"press enter when funded",
							).Show()

						pterm.Info.Println("updating dymint configuration")
						daconfig := damanager.DataLayer.GetSequencerDAConfig()
						dans := damanager.DataLayer.GetNamespaceID()

						dymintConfigPath := sequencer.GetDymintFilePath(home)
						_ = globalutils.UpdateFieldInToml(
							dymintConfigPath,
							"da_layer",
							string(rollappConfig.DA),
						)
						_ = globalutils.UpdateFieldInToml(
							dymintConfigPath,
							"namespace_id",
							dans,
						)
						_ = globalutils.UpdateFieldInToml(
							dymintConfigPath,
							"da_config",
							daconfig,
						)

						if !proceed {
							pterm.Info.Println("exiting")
							return
						}
					}
				}
			}

			// node sync
			// retrieve snapshot with the highest height

			pterm.Info.Println("done")
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

func populateSequencerMetadata(sm dymensionseqtypes.SequencerMetadata, smPath string) error {
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

	err := WriteStructToJSONFile(&sm, smPath)
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
