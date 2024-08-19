package export

import (
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	globalutils "github.com/dymensionxyz/roller/utils"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/sequencer"
	sequencerutils "github.com/dymensionxyz/roller/utils/sequencer"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Exports the current sequencer metadata into a .json file",
		Run: func(cmd *cobra.Command, args []string) {
			err := initconfig.AddFlags(cmd)
			if err != nil {
				pterm.Error.Println("failed to add flags")
				return
			}

			home, err := globalutils.ExpandHomePath(cmd.Flag(utils.FlagNames.Home).Value.String())
			if err != nil {
				pterm.Error.Println("failed to expand home directory")
				return
			}

			rollerData, err := tomlconfig.LoadRollerConfig(home)
			if err != nil {
				pterm.Error.Println("failed to load roller config file", err)
				return
			}

			hd, err := tomlconfig.LoadHubData(home)
			if err != nil {
				pterm.Error.Println("failed to load hub data from roller.toml")
			}

			rollappConfig, err := tomlconfig.LoadRollappMetadataFromChain(
				home,
				rollerData.RollappID,
				&hd,
			)
			errorhandling.PrettifyErrorIfExists(err)

			hubSeqKC := utils.KeyConfig{
				Dir:         filepath.Join(rollappConfig.Home, consts.ConfigDirName.HubKeys),
				ID:          consts.KeysIds.HubSequencer,
				ChainBinary: consts.Executables.Dymension,
				Type:        consts.SDK_ROLLAPP,
			}

			seqAddrInfo, err := utils.GetAddressInfoBinary(hubSeqKC, hubSeqKC.ChainBinary)
			if err != nil {
				pterm.Error.Println("failed to get address info: ", err)
				return
			}
			seqAddrInfo.Address = strings.TrimSpace(seqAddrInfo.Address)

			seq, err := sequencerutils.GetRegisteredSequencers(rollappConfig.RollappID, hd)
			if err != nil {
				pterm.Error.Println("failed to retrieve registered sequencers: ", err)
			}

			ok := sequencer.IsRegisteredAsSequencer(seq.Sequencers, seqAddrInfo.Address)
			if !ok {
				pterm.Error.Printf(
					"%s is not registered as a sequencer for %s\n",
					seqAddrInfo.Address,
					rollappConfig.RollappID,
				)
				return
			}

			pterm.Info.Printf(
				"%s is registered as a sequencer for %s\n",
				seqAddrInfo.Address,
				rollappConfig.RollappID,
			)
			pterm.Info.Println(
				"retrieving existing metadata",
			)

			_, err = sequencer.GetMetadata(seqAddrInfo.Address, hd)
			if err != nil {
				pterm.Error.Println("failed to retrieve metadata, ", err)
				return
			}

			pterm.Info.Println("ok")
		},
	}

	return cmd
}
