package initrollapp

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
)

var Cmd = &cobra.Command{
	Use:   "init [path-to-config-archive]",
	Short: "Inititlize RollApp locally",
	Long:  ``,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			err := initconfig.AddFlags(cmd)
			if err != nil {
				fmt.Println("failed to add flags")
				return
			}
			reader := bufio.NewReader(os.Stdin)

			fmt.Println("Do you already have rollapp config? (y/n)")
			resp, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println(err)
				return
			}

			resp = strings.TrimSpace(resp)
			resp = strings.ToLower(resp)

			if resp == "n" || resp == "no" {
				fmt.Println(
					`To generate a RollApp configuration file go to <website>
or run 'rollapp config' to expose the UI on localhost:11133.
after configuration files are generated, rerun the 'init' command`,
				)
				return
			}

			if resp == "y" || resp == "yes" {
				fmt.Println(
					"provide a path to the configuration archive file downloaded from <website>",
				)
				fp, err := reader.ReadString('\n')
				if err != nil {
					return
				}

				archivePath, err := checkConfigArchive(fp)
				if err != nil {
					fmt.Printf("failed to get archive: %v\n", err)
					return
				}

				err = runInit(cmd, archivePath)
				if err != nil {
					fmt.Printf("failed to initialize the RollApp: %v\n", err)
					return
				}

				return
			}
			return
		}

		archivePath, err := checkConfigArchive(args[0])
		if err != nil {
			fmt.Printf("failed to get archive: %v\n", err)
			return
		}

		fmt.Println(archivePath)

		err = runInit(cmd, strings.TrimSpace(archivePath))
		if err != nil {
			fmt.Printf("failed to initialize the RollApp: %v\n", err)
			return
		}
	},
}
