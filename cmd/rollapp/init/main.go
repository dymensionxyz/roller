package initrollapp

import (
	"fmt"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
)

var Cmd = &cobra.Command{
	Use:   "init [path-to-config-archive]",
	Short: "Inititlize RollApp locally",
	Long:  ``,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := initconfig.AddFlags(cmd)
		if err != nil {
			pterm.Error.Println("failed to add flags")
			return
		}

		if len(args) != 0 {
			archivePath, err := checkConfigArchive(args[0])
			if err != nil {
				fmt.Printf("failed to get archive: %v\n", err)
				return
			}

			err = runInit(cmd, WithConfig(strings.TrimSpace(archivePath)))
			if err != nil {
				fmt.Printf("failed to initialize the RollApp: %v\n", err)
				return
			}

			return
		}

		options := []string{"mock", "dymension"}
		backend, _ := pterm.DefaultInteractiveSelect.
			WithDefaultText("select the settlement layer type").
			WithOptions(options).
			Show()

		if backend == "mock" {
			err := runInit(cmd, WithMockSettlement())
			if err != nil {
				fmt.Println("failed to run init: ", err)
				return
			}
			return
		}

		hasConfig, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(
			"do you have an existing configuration archive?",
		).Show()

		if !hasConfig {
			fmt.Println(
				`To generate a RollApp configuration file go to <website>
or run 'rollapp config' to expose the UI on localhost:11133.
after configuration files are generated, rerun the 'init' command`,
			)
			return
		}

		fp, _ := pterm.DefaultInteractiveTextInput.WithDefaultText("provide the configuration archive path").
			Show()

		archivePath, err := checkConfigArchive(fp)
		if err != nil {
			fmt.Printf("failed to get archive: %v\n", err)
			return
		}

		err = runInit(cmd, WithConfig(archivePath))
		if err != nil {
			fmt.Printf("failed to initialize the RollApp: %v\n", err)
			return
		}
	},
}
