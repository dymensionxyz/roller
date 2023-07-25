package start

import (
	"fmt"
	"math/big"
	"os/exec"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"

	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"github.com/dymensionxyz/roller/relayer"
	"github.com/spf13/cobra"
)

// TODO: Test relaying on 35-C and update the prices
var oneDayRelayPriceHub = big.NewInt(1)
var oneDayRelayPriceRollapp = big.NewInt(1)

var connectionCh string

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
			logFileOption := utils.WithLogging(relayerLogFilePath)

			relayer := relayer.NewRelayer(rollappConfig.Home, rollappConfig.RollappID)
			relayer.SetLogger(*utils.GetLogger(relayerLogFilePath))

			_, _, err = relayer.LoadChannels()
			utils.PrettifyErrorIfExists(err)

			if relayer.ChannelReady() {
				fmt.Println("💈 IBC transfer channel is already established!")
			} else {
				_, err := relayer.CreateIBCChannel(logFileOption)
				utils.PrettifyErrorIfExists(err)
			}

			updateClientsCmd := relayer.GetUpdateClientsCmd()
			utils.RunCommandEvery(updateClientsCmd.Path, updateClientsCmd.Args[1:], 600, logFileOption)

			relayPacketsCmd := getRelayPacketsCmd(rollappConfig, relayer.SrcChannel)
			utils.RunCommandEvery(relayPacketsCmd.Path, relayPacketsCmd.Args[1:], 30, logFileOption)
			fmt.Printf("💈 The relayer is running successfully on you local machine! Channels: src, %s <-> %s, dst",
				relayer.SrcChannel, relayer.DstChannel)

			select {}

			// utils.RunBashCmdAsync(getRelayerStartRelaying(rollappConfig), printOutput, nil, logFileOption)
		},
	}

	return relayerStartCmd
}

func printOutput() {
	fmt.Printf("💈 The relayer is running successfully on you local machine! Channels: %s", connectionCh)
}

func getRelayPacketsCmd(config config.RollappConfig, srcChannel string) *exec.Cmd {
	return exec.Command(consts.Executables.Relayer, "tx", "relay-packets", consts.DefaultRelayerPath, srcChannel,
		"-l", "1", "--home", filepath.Join(config.Home, consts.ConfigDirName.Relayer))
}

func getRelayerStartRelaying(config config.RollappConfig) *exec.Cmd {
	return exec.Command(consts.Executables.Relayer, "start", consts.DefaultRelayerPath, "--debug-addr", "", "--home", filepath.Join(config.Home, consts.ConfigDirName.Relayer))
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
	rolRlyData, err := utils.GetRolRlyAccData(config)
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
