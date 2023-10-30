package set

import (
	"fmt"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"github.com/spf13/cobra"
	"strings"
)

var keyUpdateFuncs = map[string]func(cfg config.RollappConfig, value string) error{
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
		Short: fmt.Sprintf("Updates the specified key in all relevant places within the roller configuration files. "+
			"The Supported keys are %s", strings.Join(getSupportedKeys(), ", ")),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rlpCfg, err := config.LoadConfigFromTOML(home)
			if err != nil {
				return err
			}
			key := args[0]
			value := args[1]
			if updateFunc, exists := keyUpdateFuncs[key]; exists {
				return updateFunc(rlpCfg, value)
			}
			return fmt.Errorf("invalid key. Supported keys are: %v", strings.Join(getSupportedKeys(), ", "))
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
