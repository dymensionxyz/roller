package dependencies

import (
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/dependencies/types"
)

const (
	DefaultRelayerVersion = "v0.3.4-v2.5.2-relayer-canon-4"
	//DefaultRelayerVersion = "v0.4.3-v2.5.2-roller"
)

func DefaultRelayerPrebuiltDependencies() map[string]types.Dependency {
	return map[string]types.Dependency{
		"rly": {
			DependencyName:  "go-relayer",
			RepositoryOwner: "dymensionxyz",
			RepositoryName:  "go-relayer",
			//RepositoryUrl:   "https://github.com/artemijspavlovs/go-relayer",
			RepositoryUrl: "https://github.com/dymensionxyz/go-relayer",
			Release:       DefaultRelayerVersion,
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
