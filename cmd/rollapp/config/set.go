package config

import (
	"fmt"
	"path/filepath"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
)

func SetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "set <key> <new-value>",
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			k := args[0]
			v := args[1]
			home := cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String()

			dymintConfigPath := filepath.Join(
				home,
				consts.ConfigDirName.Rollapp,
				"config",
				"dymint.toml",
			)
			appConfigPath := filepath.Join(
				home,
				consts.ConfigDirName.Rollapp,
				"config",
				"app.toml",
			)
			// nice name, ik
			configConfigPath := filepath.Join(
				home,
				consts.ConfigDirName.Rollapp,
				"config",
				"config.toml",
			)

			// TODO: refactor, each configurable value can be a struct
			// containing config file path, key and the current value
			switch k {
			case "rollapp_minimum_gas_price":
				cfg := appConfigPath
				err := tomlconfig.UpdateFieldInFile(cfg, "minimum-gas-prices", v)
				if err != nil {
					pterm.Error.Printf("failed to update %s: %s", k, err)
					return
				}
			case "rollapp_rpc_port":
				cfg := configConfigPath
				err := tomlconfig.UpdateFieldInFile(cfg, "rpc.laddr", v)
				if err != nil {
					pterm.Error.Printf("failed to update %s: %s", k, err)
					return
				}
			case "rollapp_grpc_port":
				cfg := appConfigPath
				err := tomlconfig.UpdateFieldInFile(cfg, "grpc-web.address", v)
				if err != nil {
					pterm.Error.Printf("failed to update %s: %s", k, err)
					return
				}
			case "rollapp_rest_api_port":
				cfg := appConfigPath
				err := tomlconfig.UpdateFieldInFile(cfg, "api.address", v)
				if err != nil {
					pterm.Error.Printf("failed to update %s: %s", k, err)
					return
				}
			case "rollapp_json_rpc_port":
				cfg := appConfigPath
				err := tomlconfig.UpdateFieldInFile(cfg, "json-rpc.address", v)
				if err != nil {
					pterm.Error.Printf("failed to update %s: %s", k, err)
					return
				}
			case "rollapp_ws_port":
				cfg := appConfigPath
				err := tomlconfig.UpdateFieldInFile(cfg, "json-rpc.ws-address", v)
				if err != nil {
					pterm.Error.Printf("failed to update %s: %s", k, err)
					return
				}
			case "settlement_node_address":
				cfg := dymintConfigPath
				err := tomlconfig.UpdateFieldInFile(cfg, "settlement_node_address", v)
				if err != nil {
					pterm.Error.Printf("failed to update %s: %s", k, err)
					return
				}
			case "da_node_address":
				// Handle da_node_address
				fmt.Printf("Setting DA node address to: %s\n", v)
				// Add your logic here
			default:
				pterm.Error.Printf("unknown configuration key: %s\n", k)
				return

				// TODO: set da_config and da_layer, also the tee values
			}

			defer func() {
				pterm.Info.Println("next steps:")
				pterm.Info.Println("if this was the only configuration value you wanted to update")
				pterm.Info.Printf(
					"run %s to restart the systemd services and apply the new values\n",
					pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
						Sprintf("roller rollapp services restart"),
				)
			}()
		},
	}

	return cmd
}
