package drs

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	"github.com/dymensionxyz/roller/utils/archives"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/dependencies"
	"github.com/dymensionxyz/roller/utils/filesystem"
	firebaseutils "github.com/dymensionxyz/roller/utils/firebase"
	"github.com/dymensionxyz/roller/utils/rollapp"
	rollapputils "github.com/dymensionxyz/roller/utils/rollapp"
	"github.com/dymensionxyz/roller/utils/roller"
	"github.com/dymensionxyz/roller/utils/sequencer"
)

func UpgradeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade [drs]",
		Short: "Upgrade rollapp binary to specified drs(Dymension RollApp Standard).",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			targetDrs := args[0]
			pterm.Info.Println("preparing update")
			home, err := filesystem.ExpandHomePath(
				cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String(),
			)
			if err != nil {
				pterm.Error.Println("failed to expand home directory")
				return
			}
			rollerData, err := roller.LoadConfig(home)
			if err != nil {
				pterm.Error.Println("failed to load roller config file", err)
				return
			}

			raResp, err := rollapputils.GetMetadataFromChain(
				rollerData.RollappID,
				rollerData.HubData,
			)
			if err != nil {
				pterm.Error.Println("failed to fetch rollapp information from hub: ", err)
				return
			}

			drsVersion, err := rollapputils.ExtractDrsVersionFromBinary()
			if err != nil {
				pterm.Warning.Println("Failed to extract drs version from binary:", err)
				pterm.Info.Println("Installing the latest version", err)
			}

			if targetDrs < drsVersion {
				pterm.Error.Println("Rollapp binary DRS version already at version", drsVersion)
				return
			}

			drsInfo, err := firebaseutils.GetLatestDrsVersionCommit(
				targetDrs,
				rollerData.HubData.Environment,
			)
			if err != nil {
				pterm.Error.Println("Failed to get the latest commit for target DRS:", err)
				return
			}

			var raCommit string
			switch strings.ToLower(raResp.Rollapp.VmType) {
			case "evm":
				raCommit = drsInfo.EvmCommit
			case "wasm":
				raCommit = drsInfo.WasmCommit
			}

			if raCommit == "UNRELEASED" {
				pterm.Error.Println("rollapp does not support drs version: " + drsVersion)
				return
			}

			raNewBinDir, err := os.MkdirTemp(os.TempDir(), "rollapp-drs-update")
			if err != nil {
				pterm.Error.Println(
					"failed to create temporary directory for rollapp binary: ",
					err,
				)
				return
			}
			defer os.RemoveAll(raNewBinDir)

			rbi := dependencies.NewRollappBinaryInfo(
				raResp.Rollapp.GenesisInfo.Bech32Prefix,
				raCommit,
				strings.ToLower(raResp.Rollapp.VmType),
			)

			raDep := dependencies.DefaultRollappDependency(rbi)
			// override the binary destination with the previously created tmp directory
			// as the binary will be copied to the final destination after stopping the
			// system services
			tmpBinLocation := filepath.Join(
				raNewBinDir,
				"rollappd",
			)

			pterm.Info.Println("starting update")
			raDep.Binaries[0].BinaryDestination = tmpBinLocation
			err = dependencies.InstallBinaryFromRepo(raDep, raDep.DependencyName)
			if err != nil {
				pterm.Error.Println("failed to install rollapp binary: ", err)
				return
			}

			// repalce the current binary with the new one
			err = archives.MoveBinaryIntoPlaceAndMakeExecutable(
				tmpBinLocation,
				consts.Executables.RollappEVM,
			)
			if err != nil {
				pterm.Error.Println("failed to move rollapp binary: ", err)
				return
			}

			rollappConfig, err := rollapp.PopulateRollerConfigWithRaMetadataFromChain(
				home,
				rollerData.RollappID,
				rollerData.HubData,
			)

			damanager := datalayer.NewDAManager(
				rollappConfig.DA.Backend,
				rollappConfig.Home,
				rollappConfig.KeyringBackend,
			)

			if rollapp.IsDAConfigMigrationRequired(drsVersion, targetDrs, strings.ToLower(raResp.Rollapp.VmType)) {
				upgradeDaConfig(sequencer.GetDymintFilePath(home), *damanager)
			}

			err = tomlconfig.UpdateFieldInFile(
				filepath.Join(home, "roller.toml"),
				"rollapp_binary_version",
				raCommit,
			)
			if err != nil {
				pterm.Error.Println("failed to update rollapp binary version in config: ", err)
				return
			}

			pterm.Success.Println("update complete")
		},
	}

	return cmd
}

func upgradeDaConfig(dymintConfigPath string, daManager datalayer.DAManager) {

	daConfig := daManager.DataLayer.GetSequencerDAConfig(consts.NodeType.Sequencer)

	pterm.Info.Println("updating dymint configuration")

	_ = tomlconfig.UpdateFieldInFile(
		dymintConfigPath,
		"da_config",
		[]string{daConfig},
	)

	_ = tomlconfig.UpdateFieldInFile(
		dymintConfigPath,
		"da_layer",
		[]string{string(daManager.DaType)},
	)

}
