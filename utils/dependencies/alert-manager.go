package dependencies

import (
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/dependencies/types"
)

func DefaultAlertAgentDependency() types.Dependency {
	return types.Dependency{
		DependencyName:  "alert-agent",
		RepositoryOwner: "dymensionxyz",
		RepositoryName:  "alert-agent",
		RepositoryUrl:   "https://github.com/dymensionxyz/alert-agent",
		Release:         "v0.1.0-alpha-rc03",
		Binaries: []types.BinaryPathPair{
			{
				Binary:            "alert-agent",
				BinaryDestination: consts.Executables.AlertAgent,
			},
		},
	}
}
