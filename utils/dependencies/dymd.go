package dependencies

import (
	"errors"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/dependencies/types"
	"github.com/dymensionxyz/roller/utils/firebase"
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

func ExtractCommitFromBinaryVersion(binary string) (string, error) {
	_, err := os.Stat(binary)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	cmd := exec.Command(binary, "version", "--long")

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

func InstallCustomDymdVersion(dymdCommit string) error {
	dep := customDymdDependency(dymdCommit)

	commit, err := ExtractCommitFromBinaryVersion(consts.Executables.Dymension)
	if err != nil {
		return err
	}

	if commit == "" || commit[:6] != dep.Release[:6] {
		err := InstallBinaryFromRepo(dep, dep.DependencyName)
		if err != nil {
			return err
		}
	} else {
		pterm.Info.Println("dymd versions match, skipping installation")
	}

	return nil
}

func DefaultDymdDependency(env string) types.Dependency {
	bvi, err := firebase.GetDependencyVersions(env)
	if err != nil {
		pterm.Error.Println("failed to fetch binary versions: ", err)
		return types.Dependency{}
	}

	return types.Dependency{
		DependencyName:  "dymension",
		RepositoryOwner: "dymensionxyz",
		RepositoryName:  "dymension",
		RepositoryUrl:   "https://github.com/dymensionxyz/dymension",
		Release:         bvi.Dymd,
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
