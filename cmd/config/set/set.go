package set

import (
	"fmt"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"github.com/spf13/cobra"
)

var supportedKeys = []string{
	"rollapp-rpc-port",
}

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: fmt.Sprintf("Updates the specified key (supported keys: %v) in all relevant places within the roller configuration files.", supportedKeys),
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rlpCfg, err := config.LoadConfigFromTOML(home)
			if err != nil {
				return err
			}
			key := args[0]
			value := args[1]
			switch key {
			case "rollapp-rpc-port":
				return setRollappRPC(rlpCfg, value)
			default:
				return fmt.Errorf("invalid key. Supported keys are: %v", supportedKeys)
			}
		},
	}
	return cmd
}
