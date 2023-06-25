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
			addresses := map[string]string{}
			addresses[consts.KeyNames.DALightNode] = daAddr
			hubSeqAddr, err := utils.GetAddressBinary(utils.GetKeyConfig{
				Dir: filepath.Join(rollappConfig.Home, consts.ConfigDirName.HubKeys),
				ID:  consts.KeyNames.HubSequencer,
			}, consts.Executables.Dymension)
			utils.PrettifyErrorIfExists(err)
			addresses[consts.KeyNames.HubSequencer] = hubSeqAddr
			rollappSeqAddr, err := utils.GetAddressBinary(utils.GetKeyConfig{
				Dir: filepath.Join(rollappConfig.Home, consts.ConfigDirName.Rollapp),
				ID:  consts.KeyNames.RollappSequencer,
			}, consts.Executables.RollappEVM)
			utils.PrettifyErrorIfExists(err)
			addresses[consts.KeyNames.RollappSequencer] = rollappSeqAddr
			hubRlyAddr, err := utils.GetRelayerAddress(rollappConfig.Home, rollappConfig.HubData.ID)
			utils.PrettifyErrorIfExists(err)
			addresses[consts.KeyNames.HubRelayer] = hubRlyAddr
			rollappRlyAddr, err := utils.GetRelayerAddress(rollappConfig.Home, rollappConfig.RollappID)
			utils.PrettifyErrorIfExists(err)
			addresses[consts.KeyNames.RollappRelayer] = rollappRlyAddr
			utils.PrintAddresses(addresses)
		},
	}
	utils.AddGlobalFlags(cmd)
	return cmd
}
