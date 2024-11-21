package rollapp

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
)

func ExtractBech32PrefixFromBinary(vmType string) (string, error) {
	pterm.Info.Println("extracting bech 32 prefix")
	c := exec.Command(
		"go",
		"version",
		"-m",
		consts.Executables.RollappEVM,
	)

	out, err := bash.ExecCommandWithStdout(c)
	if err != nil {
		return "", err
	}

	lines := strings.Split(out.String(), "\n")
	var pattern string
	if vmType == "evm" {
		pattern = `github\.com/dymensionxyz/rollapp-evm/app\.AccountAddressPrefix=(\w+)`
	} else if vmType == "wasm" {
		pattern = `github\.com/dymensionxyz/rollapp-wasm/app\.AccountAddressPrefix=(\w+)`
	}
	re := regexp.MustCompile(pattern)
	var ldflags string
	var bech32Prefix string

	for _, line := range lines {
		if strings.Contains(line, "-ldflags") {
			// Print the line containing "-ldflags"
			ldflags = line
			break
		}
	}

	match := re.FindStringSubmatch(ldflags)
	if len(match) > 1 {
		// Print the captured value
		bech32Prefix = match[1]
	} else {
		return "", errors.New("rollapp binary does not contain build flags ")
	}

	return bech32Prefix, err
}

func ExtractDrsVersionFromBinary() (string, error) {
	pterm.Info.Println("extracting drs version")
	c := exec.Command(
		"go",
		"version",
		"-m",
		consts.Executables.RollappEVM,
	)

	out, err := bash.ExecCommandWithStdout(c)
	if err != nil {
		return "", err
	}

	lines := strings.Split(out.String(), "\n")
	pattern := `github\.com/dymensionxyz/dymint/version\.DrsVersion=(\w+)`

	re := regexp.MustCompile(pattern)
	var ldflags string
	var drsVersion string

	for _, line := range lines {
		if strings.Contains(line, "-ldflags") {
			// Print the line containing "-ldflags"
			ldflags = line
			fmt.Println(ldflags)
			break
		}
	}

	match := re.FindStringSubmatch(ldflags)
	if len(match) > 1 {
		// Print the captured value
		drsVersion = match[1]
	} else {
		return "", errors.New("rollapp binary does not contain build flags ")
	}

	return drsVersion, err
}
