package dependencies

import (
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/dependencies/types"
)

const (
	DefaultEibcClientVersion = "v1.0.0-alpha-rc02"
)

func DefaultEibcClientPrebuiltDependencies() map[string]types.Dependency {
	return map[string]types.Dependency{
		"eibc-client": {
			DependencyName:  "eibc-client",
			RepositoryOwner: "dymensionxyz",
			RepositoryName:  "eibc-client",
			RepositoryUrl:   "https://github.com/dymensionxyz/eibc-client",
			Release:         DefaultEibcClientVersion,
			Binaries: []types.BinaryPathPair{
				{
					Binary:            "eibc-client",
					BinaryDestination: consts.Executables.Eibc,
				},
			},
		},
		"dymd": DefaultDymdDependency(),
	}
}
