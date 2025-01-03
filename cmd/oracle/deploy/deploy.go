package deploy

import (
	"time"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/roller"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploys an oracle to the RollApp",
		Run: func(cmd *cobra.Command, args []string) {
			if err := initconfig.AddFlags(cmd); err != nil {
				pterm.Error.Printf("failed to add flags: %v\n", err)
				return
			}

			home, err := filesystem.ExpandHomePath(
				cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String(),
			)
			if err != nil {
				pterm.Error.Printf("failed to expand home directory: %v\n", err)
				return
			}

			rollerData, err := roller.LoadConfig(home)
			if err != nil {
				pterm.Error.Printf("failed to load roller config file: %v\n", err)
				return
			}

			oracle := NewOracle(rollerData)

			codeID, err := oracle.GetCodeID()
			if err != nil {
				pterm.Error.Printf("failed to get code ID: %v\n", err)
				return
			}

			if codeID == "" {
				pterm.Info.Println("no code ID found, storing contract on chain")

				if err := oracle.StoreContract(rollerData); err != nil {
					pterm.Error.Printf("failed to store contract: %v\n", err)
					return
				}

				time.Sleep(time.Second * 2)

				codeID, err = oracle.GetCodeID()
				if err != nil {
					pterm.Error.Printf("failed to get code ID: %v\n", err)
					return
				}
			}

			oracle.CodeID = codeID

			pterm.Info.Printfln("code ID: %s", codeID)

			if err := oracle.Deploy(rollerData, codeID); err != nil {
				pterm.Error.Printf("failed to deploy oracle: %v\n", err)
				return
			}
		},
	}

	return cmd
}
