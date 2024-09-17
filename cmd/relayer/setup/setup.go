package setup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	comettypes "github.com/cometbft/cometbft/types"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/relayer"
	"github.com/dymensionxyz/roller/sequencer"
	globalutils "github.com/dymensionxyz/roller/utils"
	configutils "github.com/dymensionxyz/roller/utils/config"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/config/yamlconfig"
	dymintutils "github.com/dymensionxyz/roller/utils/dymint"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/filesystem"
	genesisutils "github.com/dymensionxyz/roller/utils/genesis"
	rollapputils "github.com/dymensionxyz/roller/utils/rollapp"
)

// TODO: Test relaying on 35-C and update the prices
const (
	flagOverride = "override"
)

func Cmd() *cobra.Command {
	relayerStartCmd := &cobra.Command{
		Use:   "setup",
		Short: "Setup IBC connection between the Dymension hub and the RollApp.",
		Run: func(cmd *cobra.Command, args []string) {
			home, _ := filesystem.ExpandHomePath(cmd.Flag(utils.FlagNames.Home).Value.String())
			relayerHome := filepath.Join(home, consts.ConfigDirName.Relayer)

			genesis, err := comettypes.GenesisDocFromFile(genesisutils.GetGenesisFilePath(home))
			if err != nil {
				return
			}

			// TODO: refactor
			var need genesisutils.AppState
			j, _ := genesis.AppState.MarshalJSON()
			err = json.Unmarshal(j, &need)
			if err != nil {
				pterm.Error.Println("failed to retrieve base denom from genesis file")
				return
			}
			rollappDenom := need.Bank.Supply[0].Denom

			rollerConfigFilePath := filepath.Join(home, consts.RollerConfigFileName)
			err = globalutils.UpdateFieldInToml(rollerConfigFilePath, "base_denom", rollappDenom)
			if err != nil {
				pterm.Error.Println("failed to set base denom in roller.toml")
				return
			}

			rollappConfig, err := tomlconfig.LoadRollerConfig(home)
			if err != nil {
				pterm.Error.Printf("failed to load rollapp config: %v\n", err)
				return
			}
			relayerLogFilePath := utils.GetRelayerLogPath(rollappConfig)
			relayerLogger := utils.GetLogger(relayerLogFilePath)

			hd, err := tomlconfig.LoadHubData(home)
			if err != nil {
				pterm.Error.Println("failed to load hub data from roller.toml")
				return
			}

			rollappChainData, err := tomlconfig.LoadRollappMetadataFromChain(
				home,
				rollappConfig.RollappID,
				&hd,
			)
			errorhandling.PrettifyErrorIfExists(err)

			/* ---------------------------- Initialize relayer --------------------------- */
			defer func() {
				pterm.Debug.Println("here")
				pterm.Info.Println("reverting dymint config to 1h")
				err = dymintutils.UpdateDymintConfigForIBC(home, "1h0m0s", true)
				if err != nil {
					pterm.Error.Println("failed to update dymint config: ", err)
					return
				}
			}()

			dymintutils.WaitForHealthyRollApp("http://localhost:26657/health")
			defer func() {
				err = dymintutils.UpdateDymintConfigForIBC(home, "5s", false)
				if err != nil {
					pterm.Error.Printf("Error updating YAML: %v\n", err)
					return
				}
			}()
			outputHandler := initconfig.NewOutputHandler(false)
			isRelayerInitialized, err := filesystem.DirNotEmpty(relayerHome)
			if err != nil {
				pterm.Error.Printf("failed to check %s: %v\n", relayerHome, err)
				return
			}

			var shouldOverwrite bool
			if isRelayerInitialized {
				outputHandler.StopSpinner()
				shouldOverwrite, err = outputHandler.PromptOverwriteConfig(relayerHome)
				if err != nil {
					pterm.Error.Printf("failed to get your input: %v\n", err)
					return
				}
			}

			if shouldOverwrite {
				pterm.Info.Println("overriding the existing relayer configuration")
				err = os.RemoveAll(relayerHome)
				if err != nil {
					pterm.Error.Printf("failed to recuresively remove %s: %v\n", relayerHome, err)
					return
				}

				if runtime.GOOS == "linux" {
					pterm.Info.Println("removing old systemd services")
					for _, svc := range consts.RelayerSystemdServices {
						svcFileName := fmt.Sprintf("%s.service", svc)
						svcFilePath := filepath.Join("/etc/systemd/system/", svcFileName)

						err := filesystem.RemoveFileIfExists(svcFilePath)
						if err != nil {
							pterm.Error.Println("failed to remove systemd service: ", err)
							return
						}
					}
				}

				err = os.MkdirAll(relayerHome, 0o755)
				if err != nil {
					pterm.Error.Printf("failed to create %s: %v\n", relayerHome, err)
					return
				}
			}

			if !isRelayerInitialized || shouldOverwrite {
				// preflight checks
				blockInformation, err := rollapputils.GetCurrentHeight()
				if err != nil {
					pterm.Error.Printf("failed to get current block height: %v\n", err)
					return
				}
				currentHeight, err := strconv.Atoi(
					blockInformation.Block.Header.Height,
				)
				if err != nil {
					pterm.Error.Printf("failed to get current block height: %v\n", err)
					return
				}

				if currentHeight <= 2 {
					pterm.Warning.Println("current height is too low, updating dymint config")
					err = dymintutils.UpdateDymintConfigForIBC(home, "5s", false)
					if err != nil {
						pterm.Error.Println("failed to update dymint config: ", err)
						return
					}
				}

				rollappPrefix := rollappChainData.Bech32Prefix
				if err != nil {
					pterm.Error.Printf("failed to retrieve bech32_prefix: %v\n", err)
					return
				}

				pterm.Info.Println("initializing relayer config")
				err = initconfig.InitializeRelayerConfig(
					relayer.ChainConfig{
						ID:            rollappConfig.RollappID,
						RPC:           consts.DefaultRollappRPC,
						Denom:         rollappDenom,
						AddressPrefix: rollappPrefix,
						GasPrices:     "2000000000",
					}, relayer.ChainConfig{
						ID:            rollappConfig.HubData.ID,
						RPC:           rollappConfig.HubData.RPC_URL,
						Denom:         consts.Denoms.Hub,
						AddressPrefix: consts.AddressPrefixes.Hub,
						GasPrices:     rollappConfig.HubData.GAS_PRICE,
					}, rollappConfig,
				)
				if err != nil {
					pterm.Error.Printf(
						"failed to initialize relayer config: %v\n",
						err,
					)
					return
				}

				keys, err := initconfig.GenerateRelayerKeys(rollappConfig)
				if err != nil {
					pterm.Error.Printf("failed to create relayer keys: %v\n", err)
					return
				}

				for _, key := range keys {
					key.Print(utils.WithMnemonic(), utils.WithName())
				}

				keysToFund, err := initconfig.GetRelayerKeys(rollappConfig)
				pterm.Info.Println("please fund the hub relayer key with at least 20 dym tokens: ")
				for _, k := range keysToFund {
					k.Print(utils.WithName())
				}
				proceed, _ := pterm.DefaultInteractiveConfirm.WithDefaultValue(false).
					WithDefaultText(
						"press 'y' when the wallets are funded",
					).Show()
				if !proceed {
					return
				}

				if err != nil {
					pterm.Error.Printf("failed to create relayer keys: %v\n", err)
					return
				}

				if err := relayer.CreatePath(rollappConfig); err != nil {
					pterm.Error.Printf("failed to create relayer IBC path: %v\n", err)
					return
				}

				pterm.Info.Println("updating application relayer config")
				relayerConfigPath := filepath.Join(relayerHome, "config", "config.yaml")
				updates := map[string]interface{}{
					fmt.Sprintf("chains.%s.value.gas-adjustment", rollappConfig.HubData.ID): 1.5,
					fmt.Sprintf("chains.%s.value.gas-adjustment", rollappConfig.RollappID):  1.3,
					fmt.Sprintf("chains.%s.value.is-dym-hub", rollappConfig.HubData.ID):     true,
					fmt.Sprintf(
						"chains.%s.value.http-addr",
						rollappConfig.HubData.ID,
					): rollappConfig.HubData.API_URL,
					fmt.Sprintf("chains.%s.value.is-dym-rollapp", rollappConfig.RollappID): true,
				}
				err = yamlconfig.UpdateNestedYAML(relayerConfigPath, updates)
				if err != nil {
					pterm.Error.Printf("Error updating YAML: %v\n", err)
					return
				}

				err = dymintutils.UpdateDymintConfigForIBC(home, "5s", false)
				if err != nil {
					pterm.Error.Printf("Error updating YAML: %v\n", err)
					return
				}
			}

			if isRelayerInitialized && !shouldOverwrite {
				pterm.Info.Println("ensuring relayer keys are present")
				kc := initconfig.GetRelayerKeysConfig(rollappConfig)

				for k, v := range kc {
					pterm.Info.Printf("checking %s\n", k)

					switch v.ID {
					case consts.KeysIds.RollappRelayer:
						chainId := rollappConfig.RollappID
						isPresent, err := utils.IsRlyAddressWithNameInKeyring(v, chainId)
						if err != nil {
							pterm.Error.Printf("failed to check address: %v\n", err)
							return
						}

						if !isPresent {
							key, err := initconfig.AddRlyKey(v, rollappConfig.RollappID)
							if err != nil {
								pterm.Error.Printf("failed to add key: %v\n", err)
							}

							key.Print(utils.WithMnemonic(), utils.WithName())
						}
					case consts.KeysIds.HubRelayer:
						chainId := rollappConfig.HubData.ID
						isPresent, err := utils.IsRlyAddressWithNameInKeyring(v, chainId)
						if err != nil {
							pterm.Error.Printf("failed to check address: %v\n", err)
							return
						}
						if !isPresent {
							key, err := initconfig.AddRlyKey(v, rollappConfig.HubData.ID)
							if err != nil {
								pterm.Error.Printf("failed to add key: %v\n", err)
							}

							key.Print(utils.WithMnemonic(), utils.WithName())
						}
					default:
						pterm.Error.Println("invalid key name", err)
						return
					}
				}
			}

			if isRelayerInitialized && !shouldOverwrite {
				pterm.Info.Println("ensuring relayer keys are present")
				kc := initconfig.GetRelayerKeysConfig(rollappConfig)

				for k, v := range kc {
					pterm.Info.Printf("checking %s\n", k)

					switch v.ID {
					case consts.KeysIds.RollappRelayer:
						chainId := rollappConfig.RollappID
						isPresent, err := utils.IsRlyAddressWithNameInKeyring(v, chainId)
						if err != nil {
							pterm.Error.Printf("failed to check address: %v\n", err)
							return
						}

						if !isPresent {
							key, err := initconfig.AddRlyKey(v, rollappConfig.RollappID)
							if err != nil {
								pterm.Error.Printf("failed to add key: %v\n", err)
							}

							key.Print(utils.WithMnemonic(), utils.WithName())
						}
					case consts.KeysIds.HubRelayer:
						chainId := rollappConfig.HubData.ID
						isPresent, err := utils.IsRlyAddressWithNameInKeyring(v, chainId)
						if err != nil {
							pterm.Error.Printf("failed to check address: %v\n", err)
							return
						}
						if !isPresent {
							key, err := initconfig.AddRlyKey(v, rollappConfig.HubData.ID)
							if err != nil {
								pterm.Error.Printf("failed to add key: %v\n", err)
							}

							key.Print(utils.WithMnemonic(), utils.WithName())
						}
					default:
						pterm.Error.Println("invalid key name", err)
						return
					}
				}
			}

			err = verifyRelayerBalances(rollappConfig)
			if err != nil {
				return
			}
			rly := relayer.NewRelayer(
				rollappConfig.Home,
				rollappConfig.RollappID,
				rollappConfig.HubData.ID,
			)
			rly.SetLogger(relayerLogger)
			dymintutils.WaitForHealthyRollApp("http://localhost:26657/health")
			_, _, err = rly.LoadActiveChannel()
			if err != nil {
				pterm.Error.Printf("failed to load active channel, %v", err)
				return
			}
			logFileOption := utils.WithLoggerLogging(relayerLogger)

			// errorhandling.RequireMigrateIfNeeded(rollappConfig)

			err = rollappConfig.Validate()
			if err != nil {
				pterm.Error.Printf("failed to validate rollapp config: %v\n", err)
				return
			}

			var createIbcChannels bool

			if rly.ChannelReady() && !shouldOverwrite {
				pterm.DefaultSection.WithIndentCharacter("💈").
					Println("IBC transfer channel is already established!")

				status := fmt.Sprintf(
					"Active\nrollapp: %s\n<->\nhub: %s",
					rly.SrcChannel,
					rly.DstChannel,
				)
				err := rly.WriteRelayerStatus(status)
				if err != nil {
					fmt.Println(err)
					return
				}

				pterm.Info.Println(status)
				return
			}

			if !rly.ChannelReady() {
				createIbcChannels, _ = pterm.DefaultInteractiveConfirm.WithDefaultText(
					fmt.Sprintf(
						"no channel found. would you like to create a new IBC channel for %s?",
						rollappConfig.RollappID,
					),
				).Show()

				if !createIbcChannels {
					pterm.Warning.Println("you can't run a relayer without an ibc channel")
					return
				}
			}

			// TODO: look up relayer keys
			if createIbcChannels || shouldOverwrite {
				err = verifyRelayerBalances(rollappConfig)
				if err != nil {
					pterm.Error.Printf("failed to verify relayer balances: %v\n", err)
					return
				}

				pterm.Info.Println("establishing IBC transfer channel")
				seq := sequencer.GetInstance(rollappConfig)
				if seq == nil {
					pterm.Error.Println("failed to get sequencer sequencer instance")
					return
				}

				_, err = rly.CreateIBCChannel(shouldOverwrite, logFileOption, seq)
				if err != nil {
					pterm.Error.Printf("failed to create IBC channel: %v\n", err)
					return
				}
			}

			pterm.Info.Println("next steps:")
			pterm.Info.Printf(
				"run %s load the necessary systemd services\n",
				pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
					Sprintf("roller relayer services load"),
			)
			pterm.Warning.Println(
				"IBC channels are activated only after the first IBC transfer from RollApp to Hub",
			)
		},
	}

	relayerStartCmd.Flags().
		BoolP(flagOverride, "", false, "override the existing relayer clients and channels")
	return relayerStartCmd
}

func verifyRelayerBalances(rolCfg configutils.RollappConfig) error {
	insufficientBalances, err := relayer.GetRelayerInsufficientBalances(rolCfg)
	if err != nil {
		return err
	}
	utils.PrintInsufficientBalancesIfAny(insufficientBalances)

	return nil
}
