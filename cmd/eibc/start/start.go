package start

import (
	"fmt"
	"log"
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
				fmt.Println("eibc home directory not present, running init")
				c := eibcutils.GetInitCmd()

				_, err := bash.ExecCommandWithStdout(c)
				if err != nil {
					return
				}

				err = eibcutils.EnsureWhaleAccount()
				if err != nil {
					log.Printf("failed to create whale account: %v\n", err)
					return
				}
			}

			err = eibcutils.CreateMongoDbContainer()
			if err != nil {
				pterm.Error.Println("failed to create mongodb container:", err)
				return
			}
			pterm.Info.Println("created eibc mongodb container")

			c := eibcutils.GetStartCmd()
			err = bash.ExecCmdFollow(c)
			if err != nil {
				pterm.Error.Println("failed to create mongodb container:", err)
				return
			}
		},
	}
	return cmd
}
