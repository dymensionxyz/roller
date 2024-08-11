package set

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/utils"
	config2 "github.com/dymensionxyz/roller/utils/config"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
)

var keyUpdateFuncs = map[string]func(cfg config2.RollappConfig, value string) error{
	"rollapp-rpc-port":     setRollappRPC,
	"lc-gateway-port":      setLCGatewayPort,
	"lc-rpc-port":          setLCRPCPort,
	"rollapp-jsonrpc-port": setJsonRpcPort,
	"rollapp-ws-port":      setWSPort,
	"rollapp-grpc-port":    setGRPCPort,
	"da":                   setDA,
	"hub-rpc":              setHubRPC,
}

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "set <key> <value>",
		Short: fmt.Sprintf(
			"Updates the specified key in all relevant places within the roller configuration files. "+
				"The Supported keys are %s",
			strings.Join(getSupportedKeys(), ", "),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rlpCfg, err := tomlconfig.LoadRollerConfig(home)
			if err != nil {
				return err
			}
			key := args[0]
			value := args[1]
			if updateFunc, exists := keyUpdateFuncs[key]; exists {
				return updateFunc(rlpCfg, value)
			}
			return fmt.Errorf(
				"invalid key. Supported keys are: %v",
				strings.Join(getSupportedKeys(), ", "),
			)
		},
	}
	return cmd
}

func getSupportedKeys() []string {
	keys := make([]string, 0, len(keyUpdateFuncs))
	for key := range keyUpdateFuncs {
		keys = append(keys, key)
	}
	return keys
}
