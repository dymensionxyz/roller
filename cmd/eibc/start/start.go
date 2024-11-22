package start

import (
	"context"
	"os"
	"path/filepath"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
	eibcutils "github.com/dymensionxyz/roller/utils/eibc"
	"github.com/dymensionxyz/roller/utils/filesystem"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the eibc client",
		Run: func(cmd *cobra.Command, args []string) {
			pterm.Info.Println("starting eibc client")
			home, _ := os.UserHomeDir()
			eibcHome := filepath.Join(home, consts.ConfigDirName.Eibc)
			ok, err := filesystem.DirNotEmpty(eibcHome)
			if err != nil {
				pterm.Error.Println("failed to check eibc home directory:", err)
				return
			}

			if !ok {
				pterm.Error.Println("eibc client not initialized")
				pterm.Info.Printf(
					"run %s to initialize the eibc client\n",
					pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
						Sprintf("roller eibc init"),
				)
				return
			}

			done := make(chan error, 1)
			c := eibcutils.GetStartCmd()

			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			go func() {
				err := bash.ExecCmdFollow(
					done,
					ctx,
					c,
					nil, // No need for printOutput since we configured output above
				)

				done <- err
			}()

			select {
			case err := <-done:
				if err != nil {
					pterm.Error.Println("rollapp's process returned an error: ", err)
					return
				}
			case <-ctx.Done():
				pterm.Error.Println("context cancelled, terminating command")
				return
			}
		},
	}
	return cmd
}
