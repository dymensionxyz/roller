package dependencies

import (
	"os/exec"
	"strings"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/dependencies/types"
)

func CustomDymdDependency() types.Dependency {
	dymdCommit, _ := pterm.DefaultInteractiveTextInput.WithDefaultText(
		"provide dymensionxyz/dymension commit to build (example: 2cd612aaa6c21b473dbbb7dca9fd03b5aaae6583)",
	).Show()
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

func InstallCustomDymdVersion() error {
	dep := CustomDymdDependency()

	err := InstallBinaryFromRepo(dep, dep.DependencyName)
	if err != nil {
		return err
	}

	return nil
}

func DefaultDymdDependency() types.Dependency {
	return types.Dependency{
		DependencyName:  "dymension",
		RepositoryOwner: "dymensionxyz",
		RepositoryName:  "dymension",
		RepositoryUrl:   "https://github.com/artemijspavlovs/dymension",
		Release:         "v3.1.0-pg07",
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
