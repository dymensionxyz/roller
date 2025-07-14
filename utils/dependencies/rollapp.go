package dependencies

import (
	"fmt"
	"os/exec"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/dependencies/types"
	"github.com/dymensionxyz/roller/utils/firebase"
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

func DefaultRollappBuildableDependencies(
	raBinInfo RollappBinaryInfo,
	da string,
	env string,
) map[string]types.Dependency {
	deps := map[string]types.Dependency{}

	if da == "celestia" {
		bvi, err := firebase.GetDependencyVersions(env)
		if err != nil {
			pterm.Error.Printfln("failed to retrieve binary version for celestia light client", err)
			return nil
		}

		deps["celestia"] = CelestiaNodeDependency(*bvi)
	}
	deps["rollapp"] = DefaultRollappDependency(raBinInfo)

	return deps
}

func DefaultRollappDependency(raBinInfo RollappBinaryInfo) types.Dependency {
	var dep types.Dependency

	switch raBinInfo.VMType {
	case "evm":
		dep = types.Dependency{
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
		dep = types.Dependency{
			DependencyName:  "rollapp",
			RepositoryOwner: "dymensionxyz",
			RepositoryName:  "rollapp-wasm",
			RepositoryUrl:   "https://github.com/dymensionxyz/rollapp-wasm.git",
			Release:         raBinInfo.Commit,
			Binaries: []types.BinaryPathPair{
				{
					Binary:            "./build/rollapp-wasm",
					BinaryDestination: consts.Executables.RollappEVM,
					BuildCommand: exec.Command(
						"make",
						"build",
						fmt.Sprintf("BECH32_PREFIX=%s", raBinInfo.Bech32Prefix),
					),
				},
			},
		}
	default:
		pterm.Warning.Println("unsupported VM type")
	}

	return dep
}

func DefaultCelestiaPrebuiltDependencies() map[string]types.Dependency {
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
