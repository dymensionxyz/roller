package initrollapp

import (
	"errors"
	"os/exec"
	"regexp"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
)

const (
	// this is a constsant value of the hub genesis module
	hubgenesisAddr = "F54EBEEF798CA51615C02D13888768F9960863F2"
)

func getDebugAddrCmd(a string) *exec.Cmd {
	cmd := exec.Command(
		consts.Executables.RollappEVM,
		"debug",
		"addr",
		a,
	)

	return cmd
}

// GetHubGenesisAddress function
func GetHubGenesisAddress() (string, error) {
	cmd := getDebugAddrCmd(hubgenesisAddr)

	o, err := utils.ExecBashCommandWithStdout(cmd)
	if err != nil {
		return "", err
	}

	addr := extractBech32Acc(o.String())
	if addr == "" {
		err := errors.New("failed to extract bech 32 address")
		return "", err

	}

	return addr, nil
}

// extractBech32Acc function retrieves the bech 32 address from the output
// of 'rollappd debug addr <address>' command
func extractBech32Acc(output string) string {
	re := regexp.MustCompile(`Bech32 Acc: ([^\s]+)`)
	matches := re.FindStringSubmatch(output)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
