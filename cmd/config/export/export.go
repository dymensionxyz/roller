package export

import (
	"errors"
	"fmt"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"github.com/dymensionxyz/roller/relayer"
	"github.com/spf13/cobra"
	"math/big"
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
			rly := relayer.NewRelayer(rlpCfg.Home, rlpCfg.RollappID, rlpCfg.HubData.ID)
			srcChannel, hubChannel, err := rly.LoadChannels()
			if err != nil {
				utils.PrettifyErrorIfExists(errors.New("no relayer channels found, please create a channel before listing" +
					" your rollapp on the portal"))
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
				Ibc: IbcConfig{
					HubChannel: hubChannel,
					Channel:    srcChannel,
					Timeout:    172800000,
				},
				Type: RollApp,
				//Da:          nil,
				Description: nil,
				Analytics:   false,
			}
			if rlpCfg.VMType == config.EVM_ROLLAPP {
				evmID := config.GetEthID(rlpCfg.RollappID)
				hexEvmID, err := decimalToHexStr(evmID)
				utils.PrettifyErrorIfExists(err)
				networkJson.Evm = &EvmConfig{
					ChainId: hexEvmID,
					Rpc:     "",
				}
			}
			if rlpCfg.DA == config.Avail {
				networkJson.Da = Avail
			} else {
				networkJson.Da = Celestia
			}

			fmt.Println(rlpCfg)
		},
	}
	return cmd
}

func decimalToHexStr(decimalStr string) (string, error) {
	num := new(big.Int)
	num, ok := num.SetString(decimalStr, 10)
	if !ok {
		return "", fmt.Errorf("Failed to parse the decimal string: %s", decimalStr)
	}
	hexStr := fmt.Sprintf("0x%x", num)
	return hexStr, nil
}
