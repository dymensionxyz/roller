package dependencies

import (
	"os/exec"
	"strings"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/dependencies/types"
)

const (
	DefaultDymdCommit = "c80959d9027fae55444735386f21a9d9eeb61574"
)

func customDymdDependency(dymdCommit string) types.Dependency {
	dymdCommit = strings.TrimSpace(dymdCommit)

	return types.Dependency{
		DependencyName:  "dymension",
		RepositoryOwner: "dymensionxyz",
		RepositoryName:  "dymension",
		RepositoryUrl:   "https://github.com/dymensionxyz/dymension.git",
		Release:         dymdCommit,
		Binaries: []types.BinaryPathPair{
			{
				Binary:            "./build/dymd",
				BinaryDestination: consts.Executables.Dymension,
				BuildCommand: exec.Command(
					"make",
					"build",
				),
			},
		},
		PersistFiles: []types.PersistFile{},
	}
}

func InstallCustomDymdVersion(dymdCommit string) error {
	dep := customDymdDependency(dymdCommit)

	commit, err := ExtractCommitFromBinaryVersion(consts.Executables.Dymension)
	if err != nil {
		return err
	}

	if commit == "" || commit[:6] != dep.Release[:6] {
		err := InstallBinaryFromRepo(dep, dep.DependencyName, true)
		if err != nil {
			return err
		}
	} else {
		pterm.Info.Println("dymd versions match, skipping installation")
	}

	return nil
}

func DefaultDymdDependency() types.Dependency {
	return types.Dependency{
		DependencyName:  "dymension",
		RepositoryOwner: "dymensionxyz",
		RepositoryName:  "dymension",
		RepositoryUrl:   "https://github.com/artemijspavlovs/dymension",
		Release:         DefaultDymdCommit,
		Binaries: []types.BinaryPathPair{
			{
				Binary:            "dymd",
				BinaryDestination: consts.Executables.Dymension,
				BuildCommand:      exec.Command("make", "build"),
			},
		},
		PersistFiles: []types.PersistFile{},
	}
}
