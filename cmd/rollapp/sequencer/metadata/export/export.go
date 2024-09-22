package export

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/sequencer"
	sequencerutils "github.com/dymensionxyz/roller/utils/sequencer"
	"github.com/dymensionxyz/roller/utils/structs"
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

			home, err := filesystem.ExpandHomePath(cmd.Flag(utils.FlagNames.Home).Value.String())
			if err != nil {
				pterm.Error.Println("failed to expand home directory")
				return
			}

			rollerData, err := tomlconfig.LoadRollerConfig(home)
			if err != nil {
				pterm.Error.Println("failed to load roller config file", err)
				return
			}

			// redundant
			hd, err := tomlconfig.LoadHubData(home)
			if err != nil {
				pterm.Error.Println("failed to load hub data from roller.toml")
			}

			rollappConfig, err := tomlconfig.LoadRollappMetadataFromChain(
				home,
				rollerData.RollappID,
				&hd,
				string(rollerData.VMType),
			)
			errorhandling.PrettifyErrorIfExists(err)

			hubSeqKC := utils.KeyConfig{
				Dir:         consts.ConfigDirName.HubKeys,
				ID:          consts.KeysIds.HubSequencer,
				ChainBinary: consts.Executables.Dymension,
				Type:        consts.SDK_ROLLAPP,
			}

			seqAddrInfo, err := utils.GetAddressInfoBinary(hubSeqKC, rollappConfig.Home)
			if err != nil {
				pterm.Error.Println("failed to get address info: ", err)
				return
			}
			seqAddrInfo.Address = strings.TrimSpace(seqAddrInfo.Address)

			seq, err := sequencerutils.RegisteredRollappSequencersOnHub(rollappConfig.RollappID, hd)
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

			metadata, err := sequencer.GetMetadata(seqAddrInfo.Address, hd)
			if err != nil {
				pterm.Error.Println("failed to retrieve metadata, ", err)
				return
			}

			initDir := filepath.Join(home, consts.ConfigDirName.Rollapp, "init")

			err = os.MkdirAll(initDir, 0o755)
			if err != nil {
				pterm.Error.Println("failed to create init directory", err)
				return
			}

			metadataFilePath := filepath.Join(
				initDir, "sequencer-metadata.json",
			)
			err = structs.ExportStructToFile(
				*metadata,
				metadataFilePath,
			)
			if err != nil {
				pterm.Error.Println("failed to export metadata", err)
				return
			}

			pterm.Info.Printf("metadata successfully exported to %s\n", metadataFilePath)

			defer func() {
				pterm.Info.Println("next steps:")
				pterm.Info.Println("update the metadata file")
				pterm.Info.Printf(
					"run %s to submit a transaction to update the sequencer metadata\n",
					pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
						Sprintf("roller rollapp sequencer metadata update"),
				)
			}()
		},
	}

	return cmd
}
