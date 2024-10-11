package restart

import (
	"errors"
	"fmt"
	"runtime"
	"slices"
	"strings"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/migrations"
	"github.com/dymensionxyz/roller/utils/roller"
	"github.com/dymensionxyz/roller/utils/upgrades"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	servicemanager "github.com/dymensionxyz/roller/utils/service_manager"
)

func Cmd(services []string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Restarts the systemd services relevant to RollApp",
		Run: func(cmd *cobra.Command, args []string) {
			home, err := filesystem.ExpandHomePath(
				cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String(),
			)
			if err != nil {
				pterm.Error.Println("failed to expand home directory")
				return
			}

			err = RestartSystemdServices(services, home)
			if err != nil {
				pterm.Error.Println("failed to restart systemd services:", err)
				return
			}
		},
	}
	return cmd
}

func RestartSystemdServices(services []string, home string) error {
	if slices.Contains(services, "rollapp") {
		rollappConfig, err := roller.LoadConfig(home)
		errorhandling.PrettifyErrorIfExists(err)

		raUpgrade, err := upgrades.NewRollappUpgrade(string(rollappConfig.RollappVMType))
		if err != nil {
			pterm.Error.Println("failed to check rollapp version equality: ", err)
		}

		err = migrations.RequireRollappMigrateIfNeeded(
			raUpgrade.CurrentVersionCommit,
			rollappConfig.RollappBinaryVersion,
			string(rollappConfig.RollappVMType),
		)
		if err != nil {
			return err
		}
	}

	if runtime.GOOS == "linux" {
		for _, service := range services {
			err := servicemanager.RestartSystemdService(fmt.Sprintf("%s.service", service))
			if err != nil {
				return fmt.Errorf("failed to restart %s systemd service: %v", service, err)
			}
		}
		pterm.Success.Printf(
			"💈 Services %s restarted successfully.\n",
			strings.Join(services, ", "),
		)
	} else if runtime.GOOS == "darwin" {
		if runtime.GOOS == "linux" {
			for _, service := range services {
				err := servicemanager.RestartLaunchctlService(service)
				if err != nil {
					return fmt.Errorf("failed to restart %s systemd service: %v", service, err)
				}
			}
			pterm.Success.Printf(
				"💈 Services %s restarted successfully.\n",
				strings.Join(services, ", "),
			)
		}
	} else {
		return errors.New("os not supported")
	}
	return nil
}
