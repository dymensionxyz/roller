package show

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show the configuration of the rollapp on the local machine.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			utils.PrettifyErrorIfExists(
				printFileContent(filepath.Join(home, config.RollerConfigFileName)),
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
