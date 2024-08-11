package list

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/errorhandling"
)

var flagNames = struct {
	outputType string
}{
	outputType: "output",
}

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all the addresses of roller on the local machine.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rollappConfig, err := tomlconfig.LoadRollerConfig(home)
			errorhandling.PrettifyErrorIfExists(err)
			addresses := make([]utils.KeyInfo, 0)
			damanager := datalayer.NewDAManager(rollappConfig.DA, rollappConfig.Home)

			daAddr, err := damanager.DataLayer.GetDAAccountAddress()
			errorhandling.PrettifyErrorIfExists(err)
			if daAddr != nil {
				addresses = append(
					addresses, utils.KeyInfo{
						Address: daAddr.Address,
						Name:    damanager.GetKeyName(),
					},
				)
			}

			hubSeqInfo, err := utils.GetAddressInfoBinary(
				utils.KeyConfig{
					Dir: filepath.Join(rollappConfig.Home, consts.ConfigDirName.HubKeys),
					ID:  consts.KeysIds.HubSequencer,
				}, consts.Executables.Dymension,
			)
			errorhandling.PrettifyErrorIfExists(err)
			addresses = append(
				addresses, utils.KeyInfo{
					Address: hubSeqInfo.Address,
					Name:    consts.KeysIds.HubSequencer,
				},
			)

			raSeqInfo, err := utils.GetAddressInfoBinary(
				utils.KeyConfig{
					Dir: filepath.Join(rollappConfig.Home, consts.ConfigDirName.Rollapp),
					ID:  consts.KeysIds.RollappSequencer,
				}, rollappConfig.RollappBinary,
			)
			errorhandling.PrettifyErrorIfExists(err)
			addresses = append(
				addresses, utils.KeyInfo{
					Address: raSeqInfo.Address,
					Name:    consts.KeysIds.RollappSequencer,
				},
			)

			hubRlyAddr, err := utils.GetRelayerAddress(rollappConfig.Home, rollappConfig.HubData.ID)
			errorhandling.PrettifyErrorIfExists(err)
			addresses = append(
				addresses, utils.KeyInfo{
					Address: hubRlyAddr,
					Name:    consts.KeysIds.HubRelayer,
				},
			)

			rollappRlyAddr, err := utils.GetRelayerAddress(
				rollappConfig.Home,
				rollappConfig.RollappID,
			)
			errorhandling.PrettifyErrorIfExists(err)

			addresses = append(
				addresses, utils.KeyInfo{
					Address: rollappRlyAddr,
					Name:    consts.KeysIds.RollappRelayer,
				},
			)
			outputType := cmd.Flag(flagNames.outputType).Value.String()
			if outputType == "json" {
				errorhandling.PrettifyErrorIfExists(printAsJSON(addresses))
			} else if outputType == "text" {
				utils.PrintAddressesWithTitle(addresses)
			}
		},
	}
	cmd.Flags().StringP(flagNames.outputType, "", "text", "Output format (text|json)")
	return cmd
}

func printAsJSON(addresses []utils.KeyInfo) error {
	addrMap := make(map[string]string)
	for _, addrData := range addresses {
		addrMap[addrData.Name] = addrData.Address
	}
	data, err := json.MarshalIndent(addrMap, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling data %s", err)
	}
	fmt.Println(string(data))
	return nil
}
