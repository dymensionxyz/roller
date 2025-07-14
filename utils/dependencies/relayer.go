package dependencies

import (
	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/dependencies/types"
	"github.com/dymensionxyz/roller/utils/firebase"
)

func DefaultRelayerPrebuiltDependencies(env string) map[string]types.Dependency {
	bvi, err := firebase.GetDependencyVersions(env)
	if err != nil {
		pterm.Error.Println("failed to fetch binary versions: ", err)
		return nil
	}

	return map[string]types.Dependency{
		"rly": {
			DependencyName:  "go-relayer",
			RepositoryOwner: "dymensionxyz",
			RepositoryName:  "go-relayer",
			RepositoryUrl:   "https://github.com/dymensionxyz/go-relayer",
			Release:         bvi.Relayer,
			Binaries: []types.BinaryPathPair{
				{
					Binary:            "rly",
					BinaryDestination: consts.Executables.Relayer,
				},
			},
		},
		"dymd": DefaultDymdDependency(env),
	}
}
