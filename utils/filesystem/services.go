package filesystem

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/pterm/pterm"
)

func RemoveServiceFiles(services []string) error {
	if runtime.GOOS == "linux" {
		pterm.Info.Println("removing old systemd services")
		for _, svc := range services {
			svcFileName := fmt.Sprintf("%s.service", svc)
			svcFilePath := filepath.Join("/etc/systemd/system/", svcFileName)

			err := RemoveFileIfExists(svcFilePath)
			if err != nil {
				pterm.Error.Println("failed to remove systemd service: ", err)
				return err
			}
		}
	} else if runtime.GOOS == "darwin" {
		pterm.Info.Println("removing old systemd services")
		for _, svc := range services {
			svcFileName := fmt.Sprintf("xyz.dymension.roller.%s.plist", svc)
			svcFilePath := filepath.Join("/Library/LaunchDaemons/", svcFileName)

			err := RemoveFileIfExists(svcFilePath)
			if err != nil {
				pterm.Error.Println("failed to remove systemd service: ", err)
				return err
			}
		}
	}

	return nil
}
