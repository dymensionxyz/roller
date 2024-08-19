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
		Short: "Scale the number of fulfiller addresses to [count]",
		Long: `Scale the number of fulfiller wallets to [count]

fulfiller wallets are created to fulfill orders on behalf of the whale account

a good number to start with is 30 (default when initializing the eibc client)
`,
		Args: cobra.ExactArgs(1),
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
