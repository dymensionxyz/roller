package management_web_service

import (
	"fmt"
	"os"

	"cosmossdk.io/errors"

	"github.com/pterm/pterm"
)

// Cleanup cleans up the web service, should be called before the web service exits.
// Implementation should be panic-free.
func Cleanup() {
	if err := killEIbcClient(); err != nil {
		pterm.Error.Println("Failed to kill eIBC client:", err)
	}
}

func killEIbcClient() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic in killEIbcClient: %v", r)
		}
	}()

	pid := cache.GetEIbcClientProcessID()
	if pid == 0 {
		return
	}

	proc, errF := os.FindProcess(pid)
	if errF != nil {
		err = errors.Wrap(errF, "failed to find process")
		return
	}

	errK := proc.Kill()
	if errK != nil {
		err = errors.Wrap(errK, "failed to kill process")
		return
	}

	return nil
}
