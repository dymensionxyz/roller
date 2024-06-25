package eibc

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	global_utils "github.com/dymensionxyz/roller/utils"
)

func initCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize eibc client",
		Run: func(cmd *cobra.Command, args []string) {
			home, _ := os.UserHomeDir()
			eibcHome := filepath.Join(home, ".order-client")
			ok, err := global_utils.DirNotEmpty(eibcHome)
			if err != nil {
				return
			}

			if !ok {
				c := GetInitCommand()

				_, err = utils.ExecBashCommandWithStdout(c)
				if err != nil {
					return
				}

				err = ensureWhaleAccount()
				if err != nil {
					log.Printf("failed to create whale account: %v\n", err)
					return
				}
			}
			fmt.Println("eibc client already initialized")
		},
	}
	return cmd
}

func GetInitCommand() *exec.Cmd {
	cmd := exec.Command(
		consts.Executables.Eibc,
		"init",
	)
	return cmd
}
