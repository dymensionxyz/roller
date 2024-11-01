package dependencies

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/dependencies/types"
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

func ExtractCommitFromBinaryVersion() (string, error) {
	cmd := exec.Command(consts.Executables.Dymension, "version", "--long")

	out, err := bash.ExecCommandWithStdout(cmd)
	if err != nil {
		return "", err
	}

	lns := strings.Split(out.String(), "\n")
	var cl string
	for _, l := range lns {
		if strings.Contains(l, "commit") {
			cl = l
			break
		}
	}

	re := regexp.MustCompile(`commit: ([a-f0-9]+)`)
	match := re.FindStringSubmatch(cl)
	if len(match) > 1 {
		return match[1], nil
	} else {
		return "", errors.New("commit not found in the version output")
	}
}

func InstallCustomDymdVersion() error {
	dymdCommit, _ := pterm.DefaultInteractiveTextInput.WithDefaultText(
		"provide dymensionxyz/dymension commit to build (example: 2cd612)",
	).Show()
	dep := customDymdDependency(dymdCommit)

	commit, err := ExtractCommitFromBinaryVersion()
	if err != nil {
		return err
	}

	fmt.Println("comm", commit)
	fmt.Println("release", dep.Release)

	if commit[:6] != dep.Release[:6] {
		err := InstallBinaryFromRepo(dep, dep.DependencyName)
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
		Release:         "v3.1.0-pg10",
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
