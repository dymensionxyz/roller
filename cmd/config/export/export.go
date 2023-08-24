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
			const defaultFaucetUrl = "https://discord.com/channels/956961633165529098/1125047988247593010"
			baseDenom := "u" + rlpCfg.Denom
			coinType := 118
			if rlpCfg.VMType == config.EVM_ROLLAPP {
				coinType = 60
			}
			networkJson := NetworkJson{
				ChainId:      rlpCfg.RollappID,
				ChainName:    rlpCfg.RollappID,
				Rpc:          "",
				Rest:         "",
				Bech32Prefix: bech32,
				Currencies: []string{
					baseDenom,
				},
				NativeCurrency: baseDenom,
				StakeCurrency:  baseDenom,
				FeeCurrency:    baseDenom,
				CoinType:       coinType,
				FaucetUrl:      defaultFaucetUrl,
				Website:        "",
				Logo:           "",
				Ibc:            IbcConfig{},
				Evm:            nil,
				Type:           RollApp,
				Da:             nil,
				Description:    nil,
				Analytics:      false,
			}

			fmt.Println(rlpCfg)
		},
	}
	return cmd
}
