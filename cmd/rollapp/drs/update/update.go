package update

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/archives"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/dependencies"
	dymintutils "github.com/dymensionxyz/roller/utils/dymint"
	"github.com/dymensionxyz/roller/utils/filesystem"
	firebaseutils "github.com/dymensionxyz/roller/utils/firebase"
	rollapputils "github.com/dymensionxyz/roller/utils/rollapp"
	"github.com/dymensionxyz/roller/utils/roller"
	servicemanager "github.com/dymensionxyz/roller/utils/service_manager"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update the drs(Dymension RollApp Standard).",
		Run: func(cmd *cobra.Command, args []string) {
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

			cv, err := dependencies.ExtractCommitFromBinaryVersion(consts.Executables.RollappEVM)
			if err != nil {
				pterm.Error.Println("Failed to get the current commit:", err)
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

			drsInfo, err := firebaseutils.GetLatestDrsVersionCommit(drsVersion)
			if err != nil {
				pterm.Error.Println("Failed to get the latest commit:", err)
				return
			}

			var raCommit string
			switch strings.ToLower(raResp.Rollapp.VmType) {
			case "evm":
				raCommit = drsInfo.EvmCommit
			case "wasm":
				raCommit = drsInfo.WasmCommit
			}

			// if doesn't match, take latest as the reference
			// download the latest, build into ~/.roller/tmp
			if raCommit[:6] == cv {
				pterm.Info.Println("You are already using the latest version of DRS")
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

			dymdDep := dependencies.DefaultDymdDependency()
			pterm.Info.Println("installing dependencies")
			err = dependencies.InstallBinaryFromRelease(dymdDep)
			if err != nil {
				pterm.Error.Println("failed to install dymd: ", err)
				return
			}

			// stop services
			err = servicemanager.StopSystemServices([]string{"rollapp"})
			if err != nil {
				pterm.Error.Println("failed to stop rollapp services: ", err)
				return
			}

			// replace the current binary with the new one
			err = archives.MoveBinaryIntoPlaceAndMakeExecutable(
				tmpBinLocation,
				consts.Executables.RollappEVM,
			)
			if err != nil {
				pterm.Error.Println("failed to move rollapp binary: ", err)
				return
			}

			// start services
			err = servicemanager.StartSystemServices([]string{"rollapp"})
			if err != nil {
				pterm.Error.Println("failed to stop rollapp services: ", err)
				return
			}

			// wait for healthy endpoint
			dymintutils.WaitForHealthyRollApp("http://localhost:26657/health")

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
