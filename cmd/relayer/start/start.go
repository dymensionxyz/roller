package start

import (
	"fmt"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/spf13/cobra"
	"os/exec"
	"path/filepath"
)

func Start() *cobra.Command {
	registerCmd := &cobra.Command{
		Use:   "start",
		Short: "Starts a relayer between the Dymension hub and the rollapp.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rollappConfig, err := utils.LoadConfigFromTOML(home)
			utils.PrettifyErrorIfExists(err)
			fmt.Println(rollappConfig)
			createChannelCmd := getCreateChannelCmd(rollappConfig)
			utils.PrettifyErrorIfExists(utils.ExecBashCommand(createChannelCmd))
		},
	}
	utils.AddGlobalFlags(registerCmd)
	return registerCmd
}

func getCreateChannelCmd(rollappConfig utils.RollappConfig) *exec.Cmd {
	fmt.Println("Creating IBC channel...")
	relayerHome := filepath.Join(rollappConfig.Home, consts.ConfigDirName.Relayer)
	defaultRlyArgs := utils.GetRelayerDefaultFlags(relayerHome)
	args := []string{"transact", "link", "-t300s", consts.DefaultRelayerPath}
	args = append(args, defaultRlyArgs...)
	return exec.Command(consts.Executables.Relayer, args...)
}
