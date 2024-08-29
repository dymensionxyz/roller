package run

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	comettypes "github.com/cometbft/cometbft/types"
	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/relayer"
	"github.com/dymensionxyz/roller/sequencer"
	globalutils "github.com/dymensionxyz/roller/utils"
	"github.com/dymensionxyz/roller/utils/bash"
	configutils "github.com/dymensionxyz/roller/utils/config"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/config/yamlconfig"
	dymintutils "github.com/dymensionxyz/roller/utils/dymint"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	genesisutils "github.com/dymensionxyz/roller/utils/genesis"
	rollapputils "github.com/dymensionxyz/roller/utils/rollapp"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// TODO: Test relaying on 35-C and update the prices
const (
	flagOverride = "override"
)

// nolint gocyclo
func Cmd() *cobra.Command {
	relayerStartCmd := &cobra.Command{
		Use:   "run",
		Short: "Initialize and run a relayer between the Dymension hub and the RollApp.",
		Run: func(cmd *cobra.Command, args []string) {
			home, _ := globalutils.ExpandHomePath(cmd.Flag(utils.FlagNames.Home).Value.String())
			relayerHome := filepath.Join(home, consts.ConfigDirName.Relayer)

			genesis, err := comettypes.GenesisDocFromFile(genesisutils.GetGenesisFilePath(home))
			if err != nil {
				return
			}

			// TODO: refactor
			var need genesisutils.AppState
			j, _ := genesis.AppState.MarshalJSON()
			json.Unmarshal(j, &need)
			rollappDenom := need.Bank.Supply[0].Denom

			rollerConfigFilePath := filepath.Join(home, consts.RollerConfigFileName)
			globalutils.UpdateFieldInToml(rollerConfigFilePath, "base_denom", rollappDenom)

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
			}

			rollappChainData, err := tomlconfig.LoadRollappMetadataFromChain(
				home,
				rollappConfig.RollappID,
				&hd,
			)
			errorhandling.PrettifyErrorIfExists(err)

			/* ---------------------------- Initialize relayer --------------------------- */
			outputHandler := initconfig.NewOutputHandler(false)
			isRelayerInitialized, err := globalutils.DirNotEmpty(relayerHome)
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

				err = os.MkdirAll(relayerHome, 0o755)
				if err != nil {
					pterm.Error.Printf("failed to create %s: %v\n", relayerHome, err)
					return
				}
			}

			if !isRelayerInitialized || shouldOverwrite {
				keys, err := initconfig.GenerateRelayerKeys(rollappConfig)
				if err != nil {
					pterm.Error.Printf("failed to create relayer keys: %v\n", err)
					return
				}

				for _, key := range keys {
					key.Print(utils.WithMnemonic(), utils.WithName())
				}

				pterm.Info.Println("please fund the keys below with 20 <tokens> respectively: ")
				for _, k := range keys {
					k.Print(utils.WithName())
				}
				interactiveContinue, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(
					"Press enter when the keys are funded: ",
				).WithDefaultValue(true).Show()
				if !interactiveContinue {
					return
				}

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
						pterm.Error.Println("failed to update dymint config")
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

				pterm.Info.Println("updating application relayer config")
				path := filepath.Join(relayerHome, "config")
				data, err := os.ReadFile(filepath.Join(path, "config.yaml"))
				if err != nil {
					fmt.Printf("Error reading file: %v\n", err)
				}

				// Parse the YAML
				var node yaml.Node
				err = yaml.Unmarshal(data, &node)
				if err != nil {
					pterm.Error.Println("failed to unmarshal config.yaml")
					return
				}

				contentNode := &node
				if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
					contentNode = node.Content[0]
				}

				err = yamlconfig.UpdateNestedYAML(
					contentNode,
					[]string{"chains", rollappConfig.RollappID, "value", "gas-adjustment"},
					1.3,
				)
				if err != nil {
					fmt.Printf("Error updating YAML: %v\n", err)
					return
				}

				updatedData, err := yaml.Marshal(&node)
				if err != nil {
					fmt.Printf("Error marshaling YAML: %v\n", err)
					return
				}

				// Write the updated YAML back to the original file
				err = os.WriteFile(filepath.Join(path, "config.yaml"), updatedData, 0o644)
				if err != nil {
					fmt.Printf("Error writing file: %v\n", err)
					return
				}

				pterm.Info.Println(
					"updating dymint config to 5s block time for relayer configuration",
				)
				err = dymintutils.UpdateDymintConfigForIBC(home, "5s", false)
				if err != nil {
					pterm.Error.Println(
						"failed to update dymint config for ibc creation",
						err,
					)
					return
				}

				if err := relayer.CreatePath(rollappConfig); err != nil {
					pterm.Error.Printf("failed to create relayer IBC path: %v\n", err)
					return
				}
			}

			if isRelayerInitialized && !shouldOverwrite {
				pterm.Info.Println("ensuring relayer keys are present")
				kc := initconfig.GetRelayerKeysConfig(rollappConfig)

				for k, v := range kc {
					j, _ := json.MarshalIndent(v, "", "  ")
					fmt.Println(string(j))

					pterm.Info.Printf("checking %s in %s", k, v.Dir)
					isPresent, err := utils.IsAddressWithNameInKeyring(v, home)
					if err != nil {
						pterm.Error.Printf("failed to check address: %v", err)
					}

					if !isPresent {
						pterm.Info.Printf("%s not found in keyring, creating", k)

						var key *utils.KeyInfo
						var err error
						switch v.ID {
						case consts.KeysIds.RollappRelayer:
							key, err = initconfig.AddRlyKey(v, rollappConfig.RollappID)
						case consts.KeysIds.HubRelayer:
							key, err = initconfig.AddRlyKey(v, rollappConfig.HubData.ID)
						}
						if err != nil {
							pterm.Error.Printf("failed to create relayer key: %v", err)
						}
						key.Print(utils.WithMnemonic(), utils.WithName())
					}
				}

				err = verifyRelayerBalances(rollappConfig)
				if err != nil {
					return
				}
			}

			rly := relayer.NewRelayer(
				rollappConfig.Home,
				rollappConfig.RollappID,
				rollappConfig.HubData.ID,
			)
			rly.SetLogger(relayerLogger)
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

				status := fmt.Sprintf("Active src, %s <-> %s, dst", rly.SrcChannel, rly.DstChannel)
				err := rly.WriteRelayerStatus(status)
				if err != nil {
					fmt.Println(err)
				}
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

				_, err = rly.CreateIBCChannel(shouldOverwrite, logFileOption, seq)
				if err != nil {
					pterm.Error.Printf("failed to create IBC channel: %v\n", err)
					return
				}
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			go bash.RunCmdAsync(
				ctx,
				rly.GetStartCmd(),
				func() {},
				func(errMessage string) string { return errMessage },
				logFileOption,
			)
			pterm.Info.Printf(
				"The relayer is running successfully on you local machine!\nChannels:\nsrc, %s <-> %s, dst\n",
				rly.SrcChannel,
				rly.DstChannel,
			)

			pterm.Info.Println("reverting dymint config to 1h")
			err = dymintutils.UpdateDymintConfigForIBC(home, "1h0m0s", true)
			if err != nil {
				pterm.Error.Println("failed to update dymint config")
				return
			}

			// select {}
			pterm.Info.Println("next steps:")
			pterm.Info.Printf(
				"run %s to start rollapp and da-light-client on your local machine\n",
				pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
					Sprintf("roller relayer services load"),
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
