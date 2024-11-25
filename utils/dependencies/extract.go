package dependencies

import (
	"errors"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
)

const (
	celestiaCommitFlagPattern    = `github\.com/celestiaorg/celestia-node/nodebuilder/node\.lastCommit=(\w+)`
	celestiaAppCommitFlagPattern = `github\.com/cosmos/cosmos-sdk/version\.Commit=(\w+)`
	celKeyCommitFlagPattern      = `vcs\.revision=(\w+)`
	eibcCommitFlagPattern        = `main\.version=([\w\.-]+)`
	relayerCommitFlagPattern     = `github\.com/cosmos/relayer/v2/cmd\.Version=([\w\.-]+)`
	dymdCommitFlagPattern        = `github\.com/cosmos/cosmos-sdk/version\.Version=([\w\\.-]+)`
)

func GetCurrentCommit(binary string) (string, error) {
	switch binary {
	case consts.Executables.Dymension:
		return GetVersion(consts.Executables.Dymension)
	case consts.Executables.Celestia:
		return ExtractCommitFromBuildFlags(consts.Executables.Celestia, celestiaCommitFlagPattern)
	case consts.Executables.CelestiaApp:
		return GetVersion(consts.Executables.CelestiaApp)
	case consts.Executables.CelKey:
		return ExtractCommitFromBuildFlags(consts.Executables.CelKey, celKeyCommitFlagPattern)
	case consts.Executables.RollappEVM:
		return ExtractCommitFromBinaryVersion(consts.Executables.RollappEVM)
	case consts.Executables.Eibc:
		return ExtractCommitFromBuildFlags(consts.Executables.Eibc, eibcCommitFlagPattern)
	case consts.Executables.Relayer:
		return ExtractCommitFromBuildFlags(consts.Executables.Relayer, relayerCommitFlagPattern)
	default:
		return "", errors.New("unsupported binary")
	}
}

func ExtractCommitFromBuildFlags(binary, pattern string) (string, error) {
	c := exec.Command(
		"go",
		"version",
		"-m",
		binary,
	)

	out, err := bash.ExecCommandWithStdout(c)
	if err != nil {
		return "", err
	}

	lines := strings.Split(out.String(), "\n")

	re := regexp.MustCompile(pattern)
	var ldflags string
	var commit string

	for _, line := range lines {
		if binary == consts.Executables.CelKey {
			if strings.Contains(line, "vcs.revision") {
				ldflags = line
				break
			}
		} else {
			if strings.Contains(line, "-ldflags") {
				ldflags = line
				break
			}
		}
	}

	match := re.FindStringSubmatch(ldflags)
	if len(match) > 1 {
		commit = match[1]
	} else {
		commit = ""
	}

	return commit, err
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

func GetVersion(binary string) (string, error) {
	_, err := os.Stat(binary)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	cmd := exec.Command(binary, "version")

	out, err := bash.ExecCommandWithStdout(cmd)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(out.String()), nil
}
