package start

import (
	"fmt"
	"math/big"
	"os/exec"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"github.com/spf13/cobra"
)

// TODO: Test relaying on 35-C and update the prices
var oneDayRelayPriceHub = big.NewInt(1)
var oneDayRelayPriceRollapp = big.NewInt(1)

type RelayerConfig struct {
	SrcChannelName string
}

func Start() *cobra.Command {
	relayerStartCmd := &cobra.Command{
		Use:   "start",
		Short: "Starts a relayer between the Dymension hub and the rollapp.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rollappConfig, err := config.LoadConfigFromTOML(home)
			VerifyRelayerBalances(rollappConfig)
			utils.PrettifyErrorIfExists(err)
			relayerLogFilePath := utils.GetRelayerLogPath(rollappConfig)
			logFileOption := utils.WithLogging(relayerLogFilePath)
			srcChannelId, err := createIBCChannelIfNeeded(rollappConfig, logFileOption)
			utils.PrettifyErrorIfExists(err)
			updateClientsCmd := getUpdateClientsCmd(rollappConfig)
			utils.RunCommandEvery(updateClientsCmd.Path, updateClientsCmd.Args[1:], 60, logFileOption)
			relayPacketsCmd := getRelayPacketsCmd(rollappConfig, srcChannelId)
			utils.RunCommandEvery(relayPacketsCmd.Path, relayPacketsCmd.Args[1:], 30, logFileOption)
			fmt.Printf("💈 The relayer is running successfully on you local machine on channel %s!", srcChannelId)
			select {}
		},
	}

	return relayerStartCmd
}

func getUpdateClientsCmd(config config.RollappConfig) *exec.Cmd {
	defaultRlyArgs := getRelayerDefaultArgs(config)
	args := []string{"tx", "update-clients"}
	args = append(args, defaultRlyArgs...)
	return exec.Command(consts.Executables.Relayer, args...)
}

func getRelayPacketsCmd(config config.RollappConfig, srcChannel string) *exec.Cmd {
	return exec.Command(consts.Executables.Relayer, "tx", "relay-packets", consts.DefaultRelayerPath, srcChannel,
		"-l", "1", "--home", filepath.Join(config.Home, consts.ConfigDirName.Relayer))
}

func VerifyRelayerBalances(rolCfg config.RollappConfig) {
	insufficientBalances, err := GetRelayerInsufficientBalances(rolCfg)
	utils.PrettifyErrorIfExists(err)
	utils.PrintInsufficientBalancesIfAny(insufficientBalances)
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
	if HubRlyBalance.Cmp(oneDayRelayPriceHub) < 0 {
		insufficientBalances = append(insufficientBalances, utils.NotFundedAddressData{
			KeyName:         consts.KeysIds.HubRelayer,
			Address:         HubRlyAddr,
			CurrentBalance:  HubRlyBalance,
			RequiredBalance: oneDayRelayPriceHub,
			Denom:           consts.Denoms.Hub,
		})
	}
	return insufficientBalances, nil
}

func GetRelayerInsufficientBalances(config config.RollappConfig) ([]utils.NotFundedAddressData, error) {
	insufficientBalances, err := GetRlyHubInsufficientBalances(config)
	if err != nil {
		return insufficientBalances, err
	}
	rolRlyData, err := utils.GetRolRlyAccData(config)
	if err != nil {
		return insufficientBalances, err
	}
	if rolRlyData.Balance.Cmp(oneDayRelayPriceRollapp) < 0 {
		insufficientBalances = append(insufficientBalances, utils.NotFundedAddressData{
			KeyName:         consts.KeysIds.RollappRelayer,
			Address:         rolRlyData.Address,
			CurrentBalance:  rolRlyData.Balance,
			RequiredBalance: oneDayRelayPriceRollapp,
			Denom:           config.Denom,
		})
	}
	return insufficientBalances, nil
}
