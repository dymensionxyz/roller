package list

import (
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/spf13/cobra"
	"path/filepath"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all the addresses of roller on the local machine.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rollappConfig, err := utils.LoadConfigFromTOML(home)
			utils.PrettifyErrorIfExists(err)
			daAddr, err := utils.GetCelestiaAddress(rollappConfig.Home)
			utils.PrettifyErrorIfExists(err)
			addresses := make([]utils.AddressData, 0)
			addresses = append(addresses, utils.AddressData{
				Addr: daAddr,
				Name: consts.KeysIds.DALightNode,
			})
			hubSeqAddr, err := utils.GetAddressBinary(utils.GetKeyConfig{
				Dir: filepath.Join(rollappConfig.Home, consts.ConfigDirName.HubKeys),
				ID:  consts.KeysIds.HubSequencer,
			}, consts.Executables.Dymension)
			utils.PrettifyErrorIfExists(err)
			addresses = append(addresses, utils.AddressData{
				Addr: hubSeqAddr,
				Name: consts.KeysIds.HubSequencer,
			})
			rollappSeqAddr, err := utils.GetAddressBinary(utils.GetKeyConfig{
				Dir: filepath.Join(rollappConfig.Home, consts.ConfigDirName.Rollapp),
				ID:  consts.KeysIds.RollappSequencer,
			}, consts.Executables.RollappEVM)
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
			utils.PrintAddresses(addresses)
		},
	}
	utils.AddGlobalFlags(cmd)
	return cmd
}
