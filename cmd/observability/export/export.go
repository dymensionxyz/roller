package export

import (
	_ "embed"
	"os"
	"path/filepath"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
)

//go:embed templates/grafana/dashboard.json
var grafanaDashboardTemplate []byte

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Exports a baseline grafana dashboard for the RollApp's key metrics",
		Run: func(cmd *cobra.Command, args []string) {
			err := initconfig.AddFlags(cmd)
			if err != nil {
				pterm.Error.Println("failed to add flags")
				return
			}
			home := cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String()
			gdpath := filepath.Join(home, consts.ConfigDirName.Rollapp, "dashboard.json")

			err = os.WriteFile(gdpath, grafanaDashboardTemplate, 0o644)
			if err != nil {
				pterm.Error.Printfln("failed to export template")
				return
			}

			pterm.Info.Printf("example grafana dashboard exported to %s\n", gdpath)
		},
	}

	return cmd
}
