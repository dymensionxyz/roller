package initconfig

import (
	"fmt"
	"os"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/olekukonko/tablewriter"
)

func printInitOutput(addresses map[string]string, rollappId string) {
	fmt.Printf("💈 RollApp '%s' configuration files have been successfully generated on your local machine. Congratulations!\n\n", rollappId)

	fmt.Printf("🔑 Addresses:\n\n")

	data := [][]string{
		{"Celestia", addresses[consts.KeyNames.DALightNode]},
		{"Sequencer", addresses[consts.KeyNames.HubSequencer]},
		{"Relayer, Hub", addresses[consts.KeyNames.HubRelayer]},
		{"Relayer, RollApp", addresses[consts.KeyNames.RollappRelayer]},
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(false)
	table.AppendBulk(data)
	table.Render()

	fmt.Printf("\n🔔 Please fund these addresses to register and run the rollapp.\n")
}
