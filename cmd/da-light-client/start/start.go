package start

import (
	"fmt"
	"os/exec"
	"path/filepath"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/spf13/cobra"
)

func StartCmd() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "start",
		Short: "Runs the rollapp sequencer.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rollappConfig, err := initconfig.LoadConfigFromTOML(home)
			utils.PrettifyErrorIfExists(err)
			startRollappCmd := getCelestiaCmd(rollappConfig)
			utils.RunBashCmdAsync(startRollappCmd, printOutput, parseError)
		},
	}
	utils.AddGlobalFlags(runCmd)
	return runCmd
}

func printOutput() {
	fmt.Println("💈 The data availability light node is running on your local machine!")
	fmt.Println("💈 Light node endpoint: http://0.0.0.0:26659")
}

func parseError(errMsg string) string {
	return errMsg
}

func getCelestiaCmd(rollappConfig initconfig.InitConfig) *exec.Cmd {
	return exec.Command(
		consts.Executables.Celestia, "light", "start",
		"--core.ip", "consensus-full-arabica-8.celestia-arabica.com",
		"--node.store", filepath.Join(rollappConfig.Home, consts.ConfigDirName.DALightNode),
		"--gateway",
		"--gateway.addr", "127.0.0.1",
		"--gateway.port", "26659",
		"--p2p.network", "arabica",
	)
}
