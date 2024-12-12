package dependencies

import (
	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/dependencies/types"
	"github.com/dymensionxyz/roller/utils/firebase"
)

func DefaultEibcClientPrebuiltDependencies() map[string]types.Dependency {
	bvi, err := firebase.GetDependencyVersions()
	if err != nil {
		pterm.Error.Println("failed to fetch binary versions: ", err)
		return nil
	}

	return map[string]types.Dependency{
		"eibc-client": {
			DependencyName:  "eibc-client",
			RepositoryOwner: "dymensionxyz",
			RepositoryName:  "eibc-client",
			RepositoryUrl:   "https://github.com/dymensionxyz/eibc-client",
			Release:         bvi.EibcClient,
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
