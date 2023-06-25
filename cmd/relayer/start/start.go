package start

import (
	"fmt"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/spf13/cobra"
	"math/big"
	"os/exec"
	"path/filepath"
)

// TODO: Test relaying on 35-C and update the prices
var oneDayRelayPriceHub = big.NewInt(1)
var oneDayRelayPriceRollapp = big.NewInt(1)

type RelayerConfig struct {
	SrcChannelName string
}

func Start() *cobra.Command {
	registerCmd := &cobra.Command{
		Use:   "start",
		Short: "Starts a relayer between the Dymension hub and the rollapp.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rollappConfig, err := utils.LoadConfigFromTOML(home)
			VerifyRelayerBalances(rollappConfig)
			utils.PrettifyErrorIfExists(err)
			srcChannelId, err := createIBCChannelIfNeeded(rollappConfig)
			utils.PrettifyErrorIfExists(err)
			updateClientsCmd := getUpdateClientsCmd(rollappConfig)
			utils.RunCommandEvery(updateClientsCmd.Path, updateClientsCmd.Args[1:], 60)
			relayPacketsCmd := getRelayPacketsCmd(rollappConfig, srcChannelId)
			utils.RunCommandEvery(relayPacketsCmd.Path, relayPacketsCmd.Args[1:], 30)
			startCmd := getRlyStartCmd(rollappConfig)
			utils.RunBashCmdAsync(startCmd, func() {
				fmt.Printf("💈 The relayer is running successfully on you local machine on channel %s!", srcChannelId)
			}, parseError)
		},
	}
	utils.AddGlobalFlags(registerCmd)
	return registerCmd
}

func parseError(errStr string) string {
	// TODO
	return errStr
}

func getRlyStartCmd(config utils.RollappConfig) *exec.Cmd {
	return exec.Command(consts.Executables.Relayer, "start", consts.DefaultRelayerPath, "-l", "1", "--home",
		filepath.Join(config.Home, consts.ConfigDirName.Relayer))
}

func getUpdateClientsCmd(config utils.RollappConfig) *exec.Cmd {
	defaultRlyArgs := getRelayerDefaultArgs(config)
	args := []string{"tx", "update-clients"}
	args = append(args, defaultRlyArgs...)
	return exec.Command(consts.Executables.Relayer, args...)
}

func getRelayPacketsCmd(config utils.RollappConfig, srcChannel string) *exec.Cmd {
	return exec.Command(consts.Executables.Relayer, "tx", "relay-packets", consts.DefaultRelayerPath, srcChannel,
		"-l", "1", "--home", filepath.Join(config.Home, consts.ConfigDirName.Relayer))
}

func VerifyRelayerBalances(rolCfg utils.RollappConfig) {
	insufficientBalances, err := GetRelayerInsufficientBalances(rolCfg)
	utils.PrettifyErrorIfExists(err)
	if len(insufficientBalances) > 0 {
		utils.PrintInsufficientBalances(insufficientBalances)
	}
}

func GetRelayerInsufficientBalances(config utils.RollappConfig) ([]utils.NotFundedAddressData, error) {
	HubRlyAddr, err := utils.GetRelayerAddress(config.Home, config.HubData.ID)
	if err != nil {
		return nil, err
	}
	HubRlyBalance, err := utils.QueryBalance(utils.ChainQueryConfig{
		RPC:    config.HubData.RPC_URL,
		Denom:  consts.HubDenom,
		Binary: consts.Executables.Dymension,
	}, HubRlyAddr)
	if err != nil {
		return nil, err
	}
	insufficientBalances := make([]utils.NotFundedAddressData, 0)
	if HubRlyBalance.Cmp(oneDayRelayPriceHub) < 0 {
		insufficientBalances = append(insufficientBalances, utils.NotFundedAddressData{
			KeyName:         consts.KeyNames.HubRelayer,
			Address:         HubRlyAddr,
			CurrentBalance:  HubRlyBalance,
			RequiredBalance: oneDayRelayPriceHub,
			Denom:           consts.HubDenom,
		})
	}
	RollappRlyAddr, err := utils.GetRelayerAddress(config.Home, config.RollappID)
	if err != nil {
		return nil, err
	}
	RollappRlyBalance, err := utils.QueryBalance(utils.ChainQueryConfig{
		RPC:    consts.DefaultRollappRPC,
		Denom:  config.Denom,
		Binary: config.RollappBinary,
	}, RollappRlyAddr)
	if err != nil {
		return nil, err
	}
	if RollappRlyBalance.Cmp(oneDayRelayPriceRollapp) < 0 {
		insufficientBalances = append(insufficientBalances, utils.NotFundedAddressData{
			KeyName:         consts.KeyNames.RollappRelayer,
			Address:         RollappRlyAddr,
			CurrentBalance:  RollappRlyBalance,
			RequiredBalance: oneDayRelayPriceRollapp,
			Denom:           config.Denom,
		})
	}
	return insufficientBalances, nil
}
