package run

import (
	"fmt"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"strings"

	datalayer "github.com/dymensionxyz/roller/data_layer"
)

func NewServicesInfoTable(rollappConfig config.RollappConfig, termWidth int) *widgets.Table {
	table := widgets.NewTable()
	table.RowStyles[0] = termui.NewStyle(termui.ColorWhite, termui.ColorClear, termui.ModifierBold)
	table.SetRect(0, 13, termWidth, 22)
	table.Title = "Services Info"
	table.FillRow = true
	table.ColumnWidths = []int{termWidth / 6, termWidth / 2, termWidth / 3}
	rollappRPCPort, err := sequencer.GetRPCPort(rollappConfig)
	if err != nil {
		utils.PrettifyErrorIfExists(err)
	}
	table.Rows = [][]string{
		{"Name", "Log File", "Ports"},
		{"Sequencer", utils.GetSequencerLogPath(rollappConfig), fmt.Sprintf("%v, 8545, 1317", rollappRPCPort)},
		{"Relayer", utils.GetRelayerLogPath(rollappConfig), ""},
	}

	damanager := datalayer.NewDAManager(rollappConfig.DA, rollappConfig.Home)
	lcEndPoint, err := damanager.GetLightNodeEndpoint()
	utils.PrettifyErrorIfExists(err)
	if lcEndPoint != "" {
		parts := strings.Split(lcEndPoint, ":")
		port := parts[len(parts)-1]
		table.Rows = append(table.Rows, []string{"DA Light Client", utils.GetDALogFilePath(rollappConfig.Home), port})
	}
	return table
}
