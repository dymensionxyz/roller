package dependencies

import (
	"os/exec"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/dependencies/types"
	"github.com/dymensionxyz/roller/utils/firebase"
)

const (
	DefaultCelestiaNodeVersion = "v0.21.5"
	DefaultCelestiaAppVersion  = "v3.2.0"
)

func DefaultCelestiaNodeDependency() types.Dependency {
	return types.Dependency{
		DependencyName:  "celestia",
		RepositoryOwner: "celestiaorg",
		RepositoryName:  "celestia-node",
		RepositoryUrl:   "https://github.com/celestiaorg/celestia-node.git",
		Release:         DefaultCelestiaNodeVersion,
		Binaries: []types.BinaryPathPair{
			{
				Binary:            "./build/celestia",
				BinaryDestination: consts.Executables.Celestia,
				BuildCommand: exec.Command(
					"make",
					"build",
				),
			},
			{
				Binary:            "./cel-key",
				BinaryDestination: consts.Executables.CelKey,
				BuildCommand: exec.Command(
					"make",
					"cel-key",
				),
			},
		},
	}
}

func CelestiaNodeDependency(bvi firebase.BinaryVersionInfo) types.Dependency {
	return types.Dependency{
		DependencyName:  "celestia",
		RepositoryOwner: "celestiaorg",
		RepositoryName:  "celestia-node",
		RepositoryUrl:   "https://github.com/celestiaorg/celestia-node.git",
		Release:         bvi.CelestiaNode,
		Binaries: []types.BinaryPathPair{
			{
				Binary:            "./build/celestia",
				BinaryDestination: consts.Executables.Celestia,
				BuildCommand: exec.Command(
					"make",
					"build",
				),
			},
			{
				Binary:            "./cel-key",
				BinaryDestination: consts.Executables.CelKey,
				BuildCommand: exec.Command(
					"make",
					"cel-key",
				),
			},
		},
	}
}
