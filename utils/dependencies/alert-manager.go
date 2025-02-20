package dependencies

import (
	"os/exec"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/dependencies/types"
)

func DefaultAlertManagerDependency() types.Dependency {
	return types.Dependency{
		DependencyName:  "alert-manager",
		RepositoryOwner: "dymensionxyz",
		RepositoryName:  "alert-manager",
		RepositoryUrl:   "https://github.com/dymensionxyz/alert-manager.git",
		Release:         "v0.1.0-alpha-rc03",
		Binaries: []types.BinaryPathPair{
			{
				Binary:            "./build/alert-manager",
				BinaryDestination: consts.Executables.AlertManager,
				BuildCommand: exec.Command(
					"make",
					"build",
				),
			},
		},
	}
}
