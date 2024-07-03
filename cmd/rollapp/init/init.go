package initrollapp

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/utils"
)

var Cmd = &cobra.Command{
	Use:   "init",
	Short: "Inititlize RollApp locally",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		reader := bufio.NewReader(os.Stdin)

		fmt.Println("Do you already have rollapp config? (y/n)")
		resp, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}

		resp = strings.TrimSpace(resp)
		resp = strings.ToLower(resp)

		if resp == "y" || resp == "yes" {
			fmt.Println(
				"provide a path to the configuration archive file downloaded from <website>",
			)
			fp, err := reader.ReadString('\n')
			if err != nil {
				return
			}

			fp = strings.TrimSpace(fp)
			if fp == "" {
				fmt.Println("no path was provided")
				return
			}

			archivePath, err := expandHomePath(fp)
			if err != nil {
				return
			}

			if _, err := os.Stat(archivePath); os.IsNotExist(err) {
				fmt.Printf("the file %s does not exist. \n", fp)
				return
			}

			err = utils.ExtractZip(archivePath)
			if err != nil {
				fmt.Println("failed to extract: ", err)
			}

			return
		}

		if resp == "n" || resp == "no" {
			fmt.Println(
				`To generate a RollApp configuration file go to <website>
or run 'rollapp config' to expose the UI on localhost:11133.
After config generated, press enter..`,
			)
			return
		}

		fmt.Println("invalid or no input")
	},
}

func expandHomePath(path string) (string, error) {
	if path[:2] == "~/" {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		path = filepath.Join(usr.HomeDir, path[2:])
	}
	return path, nil
}
