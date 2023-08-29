package list

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	"github.com/spf13/cobra"
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
			rollappConfig, err := config.LoadConfigFromTOML(home)
			utils.PrettifyErrorIfExists(err)
			addresses := make([]utils.AddressData, 0)
			damanager := datalayer.NewDAManager(rollappConfig.DA, rollappConfig.Home)

			daAddr, err := damanager.DataLayer.GetDAAccountAddress()
			utils.PrettifyErrorIfExists(err)
			if daAddr != "" {
				addresses = append(addresses, utils.AddressData{
					Addr: daAddr,
					Name: damanager.GetKeyName(),
				})
			}
			hubSeqAddr, err := utils.GetAddressBinary(utils.KeyConfig{
				Dir: filepath.Join(rollappConfig.Home, consts.ConfigDirName.HubKeys),
				ID:  consts.KeysIds.HubSequencer,
			}, consts.Executables.Dymension)
			utils.PrettifyErrorIfExists(err)
			addresses = append(addresses, utils.AddressData{
				Addr: hubSeqAddr,
				Name: consts.KeysIds.HubSequencer,
			})
			rollappSeqAddr, err := utils.GetAddressBinary(utils.KeyConfig{
				Dir: filepath.Join(rollappConfig.Home, consts.ConfigDirName.Rollapp),
				ID:  consts.KeysIds.RollappSequencer,
			}, rollappConfig.RollappBinary)
			utils.PrettifyErrorIfExists(err)
			addresses = append(addresses, utils.AddressData{
				Addr: rollappSeqAddr,
				Name: consts.KeysIds.RollappSequencer,
			})
			hubRlyAddr, err := utils.GetRelayerAddress(rollappConfig.Home, rollappConfig.HubData.ID)
			utils.PrettifyErrorIfExists(err)
			addresses = append(addresses, utils.AddressData{
				Addr: hubRlyAddr,
				Name: consts.KeysIds.HubRelayer,
			})
			rollappRlyAddr, err := utils.GetRelayerAddress(rollappConfig.Home, rollappConfig.RollappID)
			utils.PrettifyErrorIfExists(err)
			addresses = append(addresses, utils.AddressData{
				Addr: rollappRlyAddr,
				Name: consts.KeysIds.RollappRelayer,
			})
			outputType := cmd.Flag(flagNames.outputType).Value.String()
			if outputType == "json" {
				utils.PrettifyErrorIfExists(printAsJSON(addresses))
			} else if outputType == "text" {
				utils.PrintAddressesWithTitle(addresses)
			}
		},
	}
	cmd.Flags().StringP(flagNames.outputType, "", "text", "Output format (text|json)")
	return cmd
}

func printAsJSON(addresses []utils.AddressData) error {
	addrMap := make(map[string]string)
	for _, addrData := range addresses {
		addrMap[addrData.Name] = addrData.Addr
	}
	data, err := json.MarshalIndent(addrMap, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling data %s", err)
	}
	fmt.Println(string(data))
	return nil
}
