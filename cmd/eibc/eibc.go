package eibc

import (
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/eibc/fulfill"
	"github.com/dymensionxyz/roller/cmd/eibc/funds"
	eibcinit "github.com/dymensionxyz/roller/cmd/eibc/init"
	"github.com/dymensionxyz/roller/cmd/eibc/scale"
	"github.com/dymensionxyz/roller/cmd/eibc/start"
	"github.com/dymensionxyz/roller/cmd/services"
	loadservices "github.com/dymensionxyz/roller/cmd/services/load"
	logservices "github.com/dymensionxyz/roller/cmd/services/logs"
	restartservices "github.com/dymensionxyz/roller/cmd/services/restart"
	startservices "github.com/dymensionxyz/roller/cmd/services/start"
	stopservices "github.com/dymensionxyz/roller/cmd/services/stop"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "eibc",
		Short: "Commands for running and managing eibc client",
	}

	cmd.AddCommand(eibcinit.Cmd())
	cmd.AddCommand(start.Cmd())
	cmd.AddCommand(scale.Cmd())
	cmd.AddCommand(funds.Cmd())
	cmd.AddCommand(fulfill.Cmd())

	sl := []string{"eibc"}
	cmd.AddCommand(
		services.Cmd(
			loadservices.Cmd(sl, cmd.Use),
			startservices.EibcCmd(),
			restartservices.Cmd(sl),
			stopservices.Cmd(sl),
			logservices.EibcCmd(),
		),
	)

	return cmd
}
