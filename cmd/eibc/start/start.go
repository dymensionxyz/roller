package start

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
	eibcutils "github.com/dymensionxyz/roller/utils/eibc"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the eibc client",
		Run: func(cmd *cobra.Command, args []string) {
			home, _ := os.UserHomeDir()
			eibcHome := filepath.Join(home, consts.ConfigDirName.Eibc)
			ok, err := filesystem.DirNotEmpty(eibcHome)
			if err != nil {
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
				return
			}

			c := eibcutils.GetStartCmd()
			err = bash.ExecCmdFollow(c)
			if err != nil {
				return
			}
		},
	}
	return cmd
}
