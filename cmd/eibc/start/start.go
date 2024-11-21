package start

import (
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

			c := eibcutils.GetStartCmd()
			err = bash.ExecCmdFollow(c, nil)
			if err != nil {
				pterm.Error.Println("failed to start the eibc client:", err)
				return
			}
		},
	}
	return cmd
}
