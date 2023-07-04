package run

import (
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"

	datalayer "github.com/dymensionxyz/roller/data_layer"
)

func NewServicesInfoTable(rollappConfig config.RollappConfig, termWidth int) *widgets.Table {
	table := widgets.NewTable()
	table.RowStyles[0] = termui.NewStyle(termui.ColorWhite, termui.ColorClear, termui.ModifierBold)
	table.SetRect(0, 13, termWidth, 22)
	table.Title = "Services Info"
	table.FillRow = true
	table.ColumnWidths = []int{termWidth / 6, termWidth / 2, termWidth / 3}
	table.Rows = [][]string{
		{"Name", "Log File", "Ports"},
		{"Sequencer", utils.GetSequencerLogPath(rollappConfig), "26657, 8545, 1317"},
		{"Relayer", utils.GetRelayerLogPath(rollappConfig), ""},
	}

	damanager := datalayer.NewDAManager(rollappConfig.DA, rollappConfig.Home)
	if damanager.GetLightNodeEndpoint() != "" {
		table.Rows = append(table.Rows, []string{"DA Light Client", utils.GetDALogFilePath(rollappConfig.Home), "26659"})
	}
	return table
}
