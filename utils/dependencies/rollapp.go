package dependencies

import (
	"fmt"
	"os/exec"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/dependencies/types"
)

type RollappBinaryInfo struct {
	Bech32Prefix string
	Commit       string
	VMType       string
}

func NewRollappBinaryInfo(bech32Prefix, commit, vmType string) RollappBinaryInfo {
	return RollappBinaryInfo{
		Bech32Prefix: bech32Prefix,
		Commit:       commit,
		VMType:       vmType,
	}
}

func DefaultRollappBuildableDependencies(raBinInfo RollappBinaryInfo) map[string]types.Dependency {
	deps := map[string]types.Dependency{}

	deps["celestia"] = DefaultCelestiaNodeDependency()

	switch raBinInfo.VMType {
	case "evm":
		deps["rollapp"] = types.Dependency{
			DependencyName:  "rollapp",
			RepositoryOwner: "dymensionxyz",
			RepositoryName:  "rollapp-evm",
			RepositoryUrl:   "https://github.com/dymensionxyz/rollapp-evm.git",
			Release:         raBinInfo.Commit,
			Binaries: []types.BinaryPathPair{
				{
					Binary:            "./build/rollapp-evm",
					BinaryDestination: consts.Executables.RollappEVM,
					BuildCommand: exec.Command(
						"make",
						"build",
						fmt.Sprintf("BECH32_PREFIX=%s", raBinInfo.Bech32Prefix),
					),
				},
			},
		}
	case "wasm":
		deps["rollapp"] = types.Dependency{
			DependencyName:  "dymd",
			RepositoryOwner: "dymensionxyz",
			RepositoryName:  "dymd",
			RepositoryUrl:   "https://github.com/dymensionxyz/dymd.git",
			Release:         DefaultDymdDependency().Release,
			Binaries: []types.BinaryPathPair{
				{
					Binary:            "./build/dymd",
					BinaryDestination: consts.Executables.Dymension,
					BuildCommand: exec.Command(
						"make",
						"build",
					),
				},
			},
		}
	default:
		pterm.Warning.Println("unsupported VM type")
	}

	return deps
}

func DefaultRollappPrebuiltDependencies() map[string]types.Dependency {
	deps := map[string]types.Dependency{
		"celestia-app": {
			DependencyName: "celestia-app",
			RepositoryUrl:  "https://github.com/celestiaorg/celestia-app",
			Release:        "v2.1.2",
			Binaries: []types.BinaryPathPair{
				{
					Binary:            "celestia-appd",
					BinaryDestination: consts.Executables.CelestiaApp,
					BuildCommand: exec.Command(
						"make",
						"build",
					),
				},
			},
		},
	}

	return deps
}
