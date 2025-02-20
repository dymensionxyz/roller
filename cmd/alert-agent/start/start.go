package start

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the Alert Agent service",
		RunE: func(cmd *cobra.Command, args []string) error {
			home := cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String()

			configDir := filepath.Join(home, consts.ConfigDirName.AlertAgent)
			if _, err := os.Stat(configDir); err != nil {
				return fmt.Errorf(
					"config directory not found, run `roller alert-agent init` first: %w",
					err,
				)
			}

			configPath := filepath.Join(configDir, "config.yaml")
			startCmd := exec.Command(
				consts.Executables.AlertAgent,
				"--config-path", configPath,
			)

			done := make(chan error, 1)
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			go func() {
				err := bash.ExecCmdFollow(
					done,
					ctx,
					startCmd,
					nil, // No need for printOutput since we configured output above
				)

				done <- err
			}()
			select {
			case err := <-done:
				if err != nil {
					return fmt.Errorf("alert-agent's process returned an error: %w", err)
				}
			case <-ctx.Done():
				return fmt.Errorf("context cancelled, terminating command")
			}

			return nil
		},
	}

	initconfig.AddGlobalFlags(cmd)
	return cmd
}
