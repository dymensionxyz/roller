package eibc

import (
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
)

func scaleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scale [count]",
		Short: "Start the eibc client",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			count := args[0]

			c := GetScaleCmd(count)

			err := bash.ExecCmdFollow(c)
			if err != nil {
				return
			}
		},
	}
	return cmd
}

func GetScaleCmd(count string) *exec.Cmd {
	cmd := exec.Command(
		consts.Executables.Eibc,
		"scale",
		count,
	)
	return cmd
}
