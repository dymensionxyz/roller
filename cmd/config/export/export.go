package export

import (
	"fmt"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export the rollapp configurations jsons needed to list your rollapp.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rlpCfg, err := config.LoadConfigFromTOML(home)
			utils.PrettifyErrorIfExists(err)
			bech32, err := getBech32Prefix(rlpCfg)
			utils.PrettifyErrorIfExists(err)
			networkJson := NetworkJson{
				ChainId:                   rlpCfg.RollappID,
				ChainName:                 rlpCfg.RollappID,
				Rpc:                       "",
				Rest:                      "",
				Bech32Prefix:              bech32,
				Currencies:                nil,
				NativeCurrency:            "",
				StakeCurrency:             "",
				FeeCurrency:               "",
				CoinType:                  0,
				FaucetUrl:                 "",
				Website:                   nil,
				ValidatorsLogosStorageDir: nil,
				Logo:                      "",
				Disabled:                  nil,
				Custom:                    nil,
				Ibc:                       IbcConfig{},
				Evm:                       nil,
				Type:                      "",
				Da:                        nil,
				Apps:                      nil,
				Description:               nil,
				IsValidator:               nil,
				Analytics:                 false,
			}

			fmt.Println(rlpCfg)
		},
	}
	return cmd
}
