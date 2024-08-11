package initrollapp

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [rollapp-id]",
		Short: "Inititlize a RollApp",
		Long:  ``,
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := initconfig.AddFlags(cmd)
			if err != nil {
				pterm.Error.Println("failed to add flags")
				return
			}

			options := []string{"mock", "dymension"}
			backend, _ := pterm.DefaultInteractiveSelect.
				WithDefaultText("select the settlement layer backend").
				WithOptions(options).
				Show()

			var raID string
			if len(args) != 0 {
				raID = args[0]
			} else {
				raID, _ = pterm.DefaultInteractiveTextInput.WithDefaultText(
					"provide a rollapp ID that you want to run the node for",
				).Show()
			}

			if backend == "mock" {
				err := runInit(cmd, "mock", raID)
				if err != nil {
					fmt.Println("failed to run init: ", err)
					return
				}
				return
			}

			envs := []string{"devnet", "testnet", "mainnet"}
			env, _ := pterm.DefaultInteractiveSelect.
				WithDefaultText("select the node type you want to run").
				WithOptions(envs).
				Show()

			err = runInit(cmd, env, raID)
			if err != nil {
				fmt.Printf("failed to initialize the RollApp: %v\n", err)
				return
			}
		},
	}
	return cmd
}
