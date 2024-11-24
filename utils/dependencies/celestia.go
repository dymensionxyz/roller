package dependencies

import (
	"os/exec"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/dependencies/types"
)

const (
	DefaultCelestiaNodeVersion = "v0.20.2-mocha"
	DefaultCelestiaAppVersion  = "v2.3.1"
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
