package dependencies

import (
	"os/exec"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/dependencies/types"
)

const (
	// 2dc9cce7b5dffc14fd4fe156d2aa9869318e0f68= "v0.20.2-mocha"
	DefaultCelestiaNodeVersion = "2dc9cce7b5dffc14fd4fe156d2aa9869318e0f68"

	// 5d6c695= "v3.0.0-mocha"
	DefaultCelestiaAppVersion = "v3.0.0-mocha"
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

func DefaultCelestiaAppPrebuiltDependency() types.Dependency {
	return types.Dependency{
		DependencyName: "celestia-app",
		RepositoryUrl:  "https://github.com/celestiaorg/celestia-app",
		Release:        DefaultCelestiaAppVersion,
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
	}
}
