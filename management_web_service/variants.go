package management_web_service

import (
	"os"
	"strings"

	"github.com/shirou/gopsutil/process"
)

var EIbcClientBinaryName = "eibc-client"

// Perform variant checks. If something wrong, call os.Exit(1) to completely terminate the entire web service.

// AnyEIbcClient checks if there is any running eIBC client.
func AnyEIbcClient() (bool, error) {
	processes, err := process.Processes()
	if err != nil {
		return false, err
	}

	for _, p := range processes {
		name, err := p.Name()
		if err != nil {
			return false, err
		}
		if name == EIbcClientBinaryName || strings.HasSuffix(name, string(os.PathSeparator)+EIbcClientBinaryName) {
			return true, nil
		}

		// Check if the process is a child process
		cmdLine, _ := p.Cmdline()
		if strings.Contains(cmdLine, EIbcClientBinaryName+" ") {
			return true, nil
		}
	}

	return false, nil
}
