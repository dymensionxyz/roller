package dependencies

import (
	"os/exec"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/dependencies/types"
)

func DefaultAlertManagerDependency() types.Dependency {
	return types.Dependency{
		DependencyName:  "alert-agent",
		RepositoryOwner: "dymensionxyz",
		RepositoryName:  "alert-agent",
		RepositoryUrl:   "https://github.com/dymensionxyz/alert-agent.git",
		Release:         "v0.1.0-alpha-rc03",
		Binaries: []types.BinaryPathPair{
			{
				Binary:            "./build/alert-agent",
				BinaryDestination: consts.Executables.AlertAgent,
				BuildCommand: exec.Command(
					"make",
					"build",
				),
			},
		},
	}
}
