package show

import (
	"fmt"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"github.com/spf13/cobra"
	"io/ioutil"
	"path/filepath"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show the configuration of the rollapp on the local machine.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			utils.PrettifyErrorIfExists(printFileContent(filepath.Join(home, config.RollerConfigFileName)))
		},
	}
	return cmd
}

func printFileContent(path string) error {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	fmt.Println(string(content))
	return nil
}
