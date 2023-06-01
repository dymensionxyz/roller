package init

import (
	"os"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

type AddressesToFund struct {
	DA           string
	HubSequencer string
	HubRelayer   string
}

func printInitOutput(addresses AddressesToFund, rollappId string) {
	color.New(color.FgCyan, color.Bold).Printf("🚀 RollApp '%s' configuration files have been successfully generated on your local machine. Congratulations!\n\n", rollappId)
	color.New(color.FgGreen, color.Bold).Printf("🔑 Key Details:\n\n")

	data := [][]string{
		{"Celestia Light Node", addresses.DA},
		{"Rollapp Hub Sequencer", addresses.HubSequencer},
		{"Rollapp Relayer", addresses.HubRelayer},
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Item", "Address"})
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(false)
	table.AppendBulk(data)
	table.Render()

	color.New(color.FgYellow, color.Bold).Printf("\n🔔 Please fund these addresses to register and run the rollapp.\n")
}
