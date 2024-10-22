package status

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/rollapp/start"
	"github.com/dymensionxyz/roller/utils/dymint"
	"github.com/dymensionxyz/roller/utils/healthagent"
	"github.com/dymensionxyz/roller/utils/roller"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show the status of the sequencer on the local machine.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String()
			rollerConfig, err := roller.LoadConfig(home)
			if err != nil {
				fmt.Println("failed to load config:", err)
				return
			}

			pidFilePath := filepath.Join(home, consts.ConfigDirName.Rollapp, "rollapp.pid")
			pid, err := os.ReadFile(pidFilePath)
			if err != nil {
				fmt.Println("failed to read pid file:", err)
				return
			}

			nodeID, err := dymint.GetNodeID(home)
			if err != nil {
				fmt.Println("failed to retrieve dymint node id:", err)
				return
			}

			ok, msg := healthagent.IsEndpointHealthy("http://localhost:26657/health")
			if !ok {
				start.PrintOutput(rollerConfig, string(pid), true, false, false, false, nodeID)
				fmt.Println("Unhealthy Message: ", msg)
				return
			}

			start.PrintOutput(rollerConfig, string(pid), true, true, true, true, nodeID)
		},
	}
	return cmd
}
