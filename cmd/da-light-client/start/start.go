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

const rpcEndpointFlag = "--rpc-endpoint"

func Cmd() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "start",
		Short: "Runs the rollapp sequencer.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rollappConfig, err := initconfig.LoadConfigFromTOML(home)
			utils.PrettifyErrorIfExists(err)
			rpcEndpoint := cmd.Flag(rpcEndpointFlag).Value.String()
			startRollappCmd := getStartCelestiaLCCmd(rollappConfig, rpcEndpoint)
			utils.RunBashCmdAsync(startRollappCmd, printOutput, parseError)
		},
	}
	utils.AddGlobalFlags(runCmd)
	addFlags(runCmd)
	return runCmd
}

func addFlags(cmd *cobra.Command) {
	cmd.Flags().StringP(rpcEndpointFlag, "", "consensus-full-arabica-8.celestia-arabica.com",
		"The DA rpc endpoint to connect to.")
}

func printOutput() {
	fmt.Println("💈 The data availability light node is running on your local machine!")
	fmt.Println("💈 Light node endpoint: http://127.0.0.1:26659")
}

func parseError(errMsg string) string {
	return errMsg
}

func getStartCelestiaLCCmd(rollappConfig initconfig.InitConfig, rpcEndpoint string) *exec.Cmd {
	return exec.Command(
		consts.Executables.Celestia, "light", "start",
		"--core.ip", rpcEndpoint,
		"--node.store", filepath.Join(rollappConfig.Home, consts.ConfigDirName.DALightNode),
		"--gateway",
		"--gateway.addr", "127.0.0.1",
		"--gateway.port", "26659",
		"--p2p.network", "arabica",
	)
}
