package rngoracle

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/roller"
)

func StartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Starts the price oracle client",
		Run: func(cmd *cobra.Command, args []string) {
			err := initconfig.AddFlags(cmd)
			if err != nil {
				pterm.Error.Println("failed to add flags")
				return
			}

			home, err := filesystem.ExpandHomePath(
				cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String(),
			)
			if err != nil {
				pterm.Error.Println("failed to expand home directory")
				return
			}

			rollerData, err := roller.LoadConfig(home)
			if err != nil {
				pterm.Error.Println("failed to load roller config: ", err)
				return
			}

			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			done := make(chan error)
			go func() {
				err := bash.ExecCmdFollow(
					done,
					ctx,
					getStartCmd(rollerData),
					nil,
				)

				done <- err
			}()

			go func() {
				err := bash.ExecCmdFollow(
					done,
					ctx,
					getStartRandomServiceCmd(),
					nil,
				)

				done <- err
			}()

			select {
			case err := <-done:
				if err != nil {
					pterm.Error.Println("rollapp's process returned an error: ", err)
					os.Exit(1)
				}
			case <-ctx.Done():
				pterm.Error.Println("context cancelled, terminating command")
				return
			}
		},
	}

	return cmd
}

func getStartCmd(rollerData roller.RollappConfig) *exec.Cmd {
	cfgPath := filepath.Join(
		rollerData.Home,
		consts.ConfigDirName.Oracle,
		consts.Oracles.Rng,
		"config.json",
	)

	args := []string{
		"start",
		"--config-path", cfgPath,
	}

	cmd := exec.Command(
		consts.Executables.RngOracle, args...,
	)
	return cmd
}

func getStartRandomServiceCmd() *exec.Cmd {
	cmd := exec.Command(
		consts.Executables.RngOracleRandomService,
	)
	return cmd
}
