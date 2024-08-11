package export

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	config2 "github.com/dymensionxyz/roller/utils/config"
	"github.com/dymensionxyz/roller/utils/config/toml"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/relayer"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export the rollapp configurations jsons needed to list your rollapp.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rlpCfg, err := toml.LoadRollerConfigFromTOML(home)
			errorhandling.PrettifyErrorIfExists(err)
			bech32, err := getBech32Prefix(rlpCfg)
			errorhandling.PrettifyErrorIfExists(err)
			faucetUrls := map[string]string{
				consts.LocalHubID:   "",
				consts.TestnetHubID: "https://discord.com/channels/956961633165529098/1196803789911498763",
				consts.MainnetHubID: "",
			}
			baseDenom := rlpCfg.Denom

			coinType := 118
			if rlpCfg.VMType == consts.EVM_ROLLAPP {
				coinType = 60
			}
			rly := relayer.NewRelayer(rlpCfg.Home, rlpCfg.RollappID, rlpCfg.HubData.ID)
			_, _, err = rly.LoadActiveChannel()
			if err != nil || rly.SrcChannel == "" || rly.DstChannel == "" {
				errorhandling.PrettifyErrorIfExists(
					errors.New(
						"failed to export rollapp json." +
							" Please verify that the rollapp is running on your local machine and a relayer channel has been established",
					),
				)
			}
			logoDefaultPath := fmt.Sprintf("/logos/%s.png", rlpCfg.RollappID)
			networkJson := NetworkJson{
				ChainId:      rlpCfg.RollappID,
				ChainName:    rlpCfg.RollappID,
				Rpc:          "",
				Rest:         "",
				Bech32Prefix: bech32,
				Currencies: []Currency{
					{
						DisplayDenom: baseDenom[1:],
						BaseDenom:    baseDenom,
						Decimals:     rlpCfg.Decimals,
						Logo:         logoDefaultPath,
						CurrencyType: "main",
					},
				},
				CoinType:  coinType,
				FaucetUrl: faucetUrls[rlpCfg.HubData.ID],
				Website:   "",
				Logo:      logoDefaultPath,
				Ibc: IbcConfig{
					HubChannel: rly.SrcChannel,
					Channel:    rly.DstChannel,
					Timeout:    172800000,
				},
				Type:      RollApp,
				Analytics: true,
			}
			if rlpCfg.VMType == consts.EVM_ROLLAPP {
				evmID := config2.GetEthID(rlpCfg.RollappID)
				hexEvmID, err := decimalToHexStr(evmID)
				errorhandling.PrettifyErrorIfExists(err)
				networkJson.Evm = &EvmConfig{
					ChainId: hexEvmID,
					Rpc:     "",
				}
			}
			if rlpCfg.DA == consts.Avail {
				networkJson.Da = Avail
			} else {
				networkJson.Da = Celestia
			}

			networkJsonString, err := json.MarshalIndent(networkJson, "", "  ")
			errorhandling.PrettifyErrorIfExists(err)
			println("💈 networks.json:")
			println(string(networkJsonString))
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
