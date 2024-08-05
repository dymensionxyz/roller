package run

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strconv"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"github.com/dymensionxyz/roller/relayer"
	"github.com/dymensionxyz/roller/sequencer"
	globalutils "github.com/dymensionxyz/roller/utils"
	dymintutils "github.com/dymensionxyz/roller/utils/dymint"
	rollapputils "github.com/dymensionxyz/roller/utils/rollapp"
)

// TODO: Test relaying on 35-C and update the prices
var (
	oneDayRelayPriceHub     = big.NewInt(1)
	oneDayRelayPriceRollapp = big.NewInt(1)
)

const (
	flagOverride = "override"
)

// nolint gocyclo
func Cmd() *cobra.Command {
	relayerStartCmd := &cobra.Command{
		Use:   "run",
		Short: "Runs a relayer between the Dymension hub and the rollapp.",
		Run: func(cmd *cobra.Command, args []string) {
			home, _ := globalutils.ExpandHomePath(cmd.Flag(utils.FlagNames.Home).Value.String())
			relayerHome := filepath.Join(home, consts.ConfigDirName.Relayer)
			rollappConfig, err := config.LoadRollerConfigFromTOML(home)
			if err != nil {
				pterm.Error.Printf("failed to load rollapp config: %v\n", err)
				return
			}
			rollerConfigFilePath := filepath.Join(home, "roller.toml")
			relayerLogFilePath := utils.GetRelayerLogPath(rollappConfig)
			relayerLogger := utils.GetLogger(relayerLogFilePath)

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
				// preflight checks
				blockInformation, err := rollapputils.GetCurrentHeight()
				if err != nil {
					fmt.Println("failed to retrieve current height")
				}
				currentHeight, err := strconv.Atoi(
					strconv.FormatInt(blockInformation.Block.Header.Height, 10),
				)
				if err != nil {
					fmt.Println("failed to retrieve current height")
				}
				if currentHeight <= 2 {
					pterm.Warning.Println("current height is too low, updating dymint config")
					err := dymintutils.UpdateDymintConfigForIBC(home)
					if err != nil {
						pterm.Error.Println("failed to update dymint config")
						return
					}
				}

				rollappPrefix, err := globalutils.GetKeyFromTomlFile(
					rollerConfigFilePath,
					"bech32_prefix",
				)
				if err != nil {
					pterm.Error.Printf("failed to retrieve bech32_prefix: %v\n", err)
					return
				}

				err = initconfig.InitializeRelayerConfig(
					relayer.ChainConfig{
						ID:            rollappConfig.RollappID,
						RPC:           consts.DefaultRollappRPC,
						Denom:         rollappConfig.Denom,
						AddressPrefix: rollappPrefix,
						GasPrices:     "0",
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

				err = dymintutils.UpdateDymintConfigForIBC(home)
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

			utils.RequireMigrateIfNeeded(rollappConfig)

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
				if shouldOverwrite {
					keys, err := initconfig.GenerateRelayerKeys(rollappConfig)
					if err != nil {
						pterm.Error.Printf("failed to create relayer keys: %v\n", err)
						return
					}

					for _, key := range keys {
						key.Print(utils.WithMnemonic(), utils.WithName())
					}

					pterm.Info.Println("please fund the keys below with X <tokens> respectively: ")
					for _, k := range keys {
						k.Print(utils.WithName())
					}
					interactiveContinue, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(
						"Press enter when the keys are funded: ",
					).WithDefaultValue(true).Show()
					if !interactiveContinue {
						return
					}
				}

				err = VerifyRelayerBalances(rollappConfig)
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
			go utils.RunBashCmdAsync(
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

			select {}
		},
	}

	relayerStartCmd.Flags().
		BoolP(flagOverride, "", false, "override the existing relayer clients and channels")
	return relayerStartCmd
}

func VerifyRelayerBalances(rolCfg config.RollappConfig) error {
	insufficientBalances, err := GetRelayerInsufficientBalances(rolCfg)
	if err != nil {
		return err
	}
	utils.PrintInsufficientBalancesIfAny(insufficientBalances)

	return nil
}

func GetRlyHubInsufficientBalances(
	config config.RollappConfig,
) ([]utils.NotFundedAddressData, error) {
	HubRlyAddr, err := utils.GetRelayerAddress(config.Home, config.HubData.ID)
	if err != nil {
		pterm.Error.Printf("failed to get relayer address: %v", err)
		return nil, err
	}

	HubRlyBalance, err := utils.QueryBalance(
		utils.ChainQueryConfig{
			RPC:    config.HubData.RPC_URL,
			Denom:  consts.Denoms.Hub,
			Binary: consts.Executables.Dymension,
		}, HubRlyAddr,
	)
	if err != nil {
		pterm.Error.Printf("failed to query %s balances: %v", HubRlyAddr, err)
		return nil, err
	}

	insufficientBalances := make([]utils.NotFundedAddressData, 0)
	if HubRlyBalance.Amount.Cmp(oneDayRelayPriceHub) < 0 {
		insufficientBalances = append(
			insufficientBalances, utils.NotFundedAddressData{
				KeyName:         consts.KeysIds.HubRelayer,
				Address:         HubRlyAddr,
				CurrentBalance:  HubRlyBalance.Amount,
				RequiredBalance: oneDayRelayPriceHub,
				Denom:           consts.Denoms.Hub,
				Network:         config.HubData.ID,
			},
		)
	}
	return insufficientBalances, nil
}

func GetRelayerInsufficientBalances(
	config config.RollappConfig,
) ([]utils.NotFundedAddressData, error) {
	insufficientBalances, err := GetRlyHubInsufficientBalances(config)
	if err != nil {
		return insufficientBalances, err
	}

	rolRlyData, err := relayer.GetRolRlyAccData(config)
	if err != nil {
		return insufficientBalances, err
	}

	if rolRlyData.Balance.Amount.Cmp(oneDayRelayPriceRollapp) < 0 {
		insufficientBalances = append(
			insufficientBalances, utils.NotFundedAddressData{
				KeyName:         consts.KeysIds.RollappRelayer,
				Address:         rolRlyData.Address,
				CurrentBalance:  rolRlyData.Balance.Amount,
				RequiredBalance: oneDayRelayPriceRollapp,
				Denom:           config.Denom,
				Network:         config.RollappID,
			},
		)
	}

	return insufficientBalances, nil
}
