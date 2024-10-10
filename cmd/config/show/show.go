package show

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/errorhandling"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show the configuration of the rollapp on the local machine.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String()
			errorhandling.PrettifyErrorIfExists(
				printFileContent(filepath.Join(home, consts.RollerConfigFileName)),
			)
		},
	}
	return cmd
}

func printFileContent(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	fmt.Println(string(content))
	return nil
}
