package dependencies

import (
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/dependencies/types"
)

func DefaultRelayerPrebuiltDependencies() map[string]types.Dependency {
	return map[string]types.Dependency{
		"rly": {
			DependencyName:  "go-relayer",
			RepositoryOwner: "artemijspavlovs",
			RepositoryName:  "go-relayer",
			RepositoryUrl:   "https://github.com/artemijspavlovs/go-relayer",
			Release:         "v0.4.0-v2.5.2-relayer-pg-roller",
			Binaries: []types.BinaryPathPair{
				{
					Binary:            "rly",
					BinaryDestination: consts.Executables.Relayer,
				},
			},
		},
		"dymd": DefaultDymdDependency(),
	}
}
