package start

import (
	"fmt"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/spf13/cobra"
	"os/exec"
	"path/filepath"
)

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
