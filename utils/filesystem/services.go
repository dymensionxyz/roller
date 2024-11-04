package filesystem

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/pterm/pterm"
)

func RemoveServiceFiles(services []string) error {
	pterm.Info.Println("removing old systemd services")

	switch runtime.GOOS {
	case "linux":
		for _, svc := range services {
			svcFileName := fmt.Sprintf("%s.service", svc)
			svcFilePath := filepath.Join("/etc/systemd/system/", svcFileName)

			err := RemoveFileIfExists(svcFilePath)
			if err != nil {
				pterm.Error.Println("failed to remove systemd service: ", err)
				return err
			}
		}
	case "darwin":
		for _, svc := range services {
			svcFileName := fmt.Sprintf("xyz.dymension.roller.%s.plist", svc)
			svcFilePath := filepath.Join("/Library/LaunchDaemons/", svcFileName)

			err := RemoveFileIfExists(svcFilePath)
			if err != nil {
				pterm.Error.Println("failed to remove systemd service: ", err)
				return err
			}
		}
	default:
		pterm.Error.Println("OS not supported")
	}

	return nil
}
