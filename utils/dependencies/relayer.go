package dependencies

import (
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/dependencies/types"
)

const (
	DefaultRelayerVersion = "v0.4.1-v2.5.2-roller"
)

func DefaultRelayerPrebuiltDependencies() map[string]types.Dependency {
	return map[string]types.Dependency{
		"rly": {
			DependencyName:  "go-relayer",
			RepositoryOwner: "artemijspavlovs",
			RepositoryName:  "go-relayer",
			RepositoryUrl:   "https://github.com/artemijspavlovs/go-relayer",
			Release:         DefaultRelayerVersion,
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
