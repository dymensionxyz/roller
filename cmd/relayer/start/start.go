package start

import (
	"fmt"
	"math/big"

	"github.com/dymensionxyz/roller/cmd/consts"

	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"github.com/dymensionxyz/roller/relayer"
	"github.com/spf13/cobra"
)

// TODO: Test relaying on 35-C and update the prices
var oneDayRelayPriceHub = big.NewInt(1)
var oneDayRelayPriceRollapp = big.NewInt(1)

const (
	flagOverride = "override"
)

func Start() *cobra.Command {
	relayerStartCmd := &cobra.Command{
		Use:   "start",
		Short: "Starts a relayer between the Dymension hub and the rollapp.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rollappConfig, err := config.LoadConfigFromTOML(home)
			utils.PrettifyErrorIfExists(err)

			VerifyRelayerBalances(rollappConfig)
			relayerLogFilePath := utils.GetRelayerLogPath(rollappConfig)
			logger := utils.GetLogger(relayerLogFilePath)
			logFileOption := utils.WithLoggerLogging(logger)
			rly := relayer.NewRelayer(rollappConfig.Home, rollappConfig.RollappID, rollappConfig.HubData.ID)
			rly.SetLogger(logger)

			_, _, err = rly.LoadChannels()
			utils.PrettifyErrorIfExists(err)

			override := cmd.Flag(flagOverride).Changed
			if override {
				fmt.Println("💈 Overriding the existing relayer channel")
			}
			if rly.ChannelReady() && !override {
				fmt.Println("💈 IBC transfer channel is already established!")
			} else {
				fmt.Println("💈 Establishing IBC transfer channel")
				_, err := rly.CreateIBCChannel(override, logFileOption)
				utils.PrettifyErrorIfExists(err)
			}

			updateClientsCmd := rly.GetUpdateClientsCmd()
			utils.RunCommandEvery(updateClientsCmd.Path, updateClientsCmd.Args[1:], 5, logFileOption)
			relayPacketsCmd := rly.GetRelayPacketsCmd()
			utils.RunCommandEvery(relayPacketsCmd.Path, relayPacketsCmd.Args[1:], 5, logFileOption)
			relayAcksCmd := rly.GetRelayAcksCmd()
			utils.RunCommandEvery(relayAcksCmd.Path, relayAcksCmd.Args[1:], 5, logFileOption)
			fmt.Printf("💈 The relayer is running successfully on you local machine! Channels: src, %s <-> %s, dst",
				rly.SrcChannel, rly.DstChannel)

			select {}
		},
	}

	relayerStartCmd.Flags().BoolP(flagOverride, "", false, "override the existing relayer clients and channels")
	return relayerStartCmd
}

func VerifyRelayerBalances(rolCfg config.RollappConfig) {
	insufficientBalances, err := GetRelayerInsufficientBalances(rolCfg)
	utils.PrettifyErrorIfExists(err)
	utils.PrintInsufficientBalancesIfAny(insufficientBalances, rolCfg)
}

func GetRlyHubInsufficientBalances(config config.RollappConfig) ([]utils.NotFundedAddressData, error) {
	HubRlyAddr, err := utils.GetRelayerAddress(config.Home, config.HubData.ID)
	if err != nil {
		return nil, err
	}
	HubRlyBalance, err := utils.QueryBalance(utils.ChainQueryConfig{
		RPC:    config.HubData.RPC_URL,
		Denom:  consts.Denoms.Hub,
		Binary: consts.Executables.Dymension,
	}, HubRlyAddr)
	if err != nil {
		return nil, err
	}
	insufficientBalances := make([]utils.NotFundedAddressData, 0)
	if HubRlyBalance.Amount.Cmp(oneDayRelayPriceHub) < 0 {
		insufficientBalances = append(insufficientBalances, utils.NotFundedAddressData{
			KeyName:         consts.KeysIds.HubRelayer,
			Address:         HubRlyAddr,
			CurrentBalance:  HubRlyBalance.Amount,
			RequiredBalance: oneDayRelayPriceHub,
			Denom:           consts.Denoms.Hub,
			Network:         config.HubData.ID,
		})
	}
	return insufficientBalances, nil
}

func GetRelayerInsufficientBalances(config config.RollappConfig) ([]utils.NotFundedAddressData, error) {
	insufficientBalances, err := GetRlyHubInsufficientBalances(config)
	if err != nil {
		return insufficientBalances, err
	}
	rolRlyData, err := relayer.GetRolRlyAccData(config)
	if err != nil {
		return insufficientBalances, err
	}
	if rolRlyData.Balance.Amount.Cmp(oneDayRelayPriceRollapp) < 0 {
		insufficientBalances = append(insufficientBalances, utils.NotFundedAddressData{
			KeyName:         consts.KeysIds.RollappRelayer,
			Address:         rolRlyData.Address,
			CurrentBalance:  rolRlyData.Balance.Amount,
			RequiredBalance: oneDayRelayPriceRollapp,
			Denom:           config.Denom,
			Network:         config.RollappID,
		})
	}
	return insufficientBalances, nil
}
