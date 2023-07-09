package run

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/dymensionxyz/roller/config"
	servicemanager "github.com/dymensionxyz/roller/utils/service_manager"
	"github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

func initializeUI() {
	if err := termui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
}

func buildTitleParagraph(rollappConfig config.RollappConfig, termWidth int) *widgets.Paragraph {
	p := widgets.NewParagraph()
	p.Text = fmt.Sprintf("💈 Rollapp '%s' is successfully running on your local machine, connected to Dymension hub '%s'.",
		rollappConfig.RollappID, rollappConfig.HubData.ID)
	p.SetRect(0, 0, termWidth, 3)
	return p
}

func NewServiceStatusTable(termWidth int) *widgets.Table {
	table := widgets.NewTable()
	table.RowStyles[0] = termui.NewStyle(termui.ColorWhite, termui.ColorClear, termui.ModifierBold)
	table.SetRect(0, 3, termWidth, 12)
	table.Title = "Services Status"
	return table
}

func updateUITable(oldUIData, serviceData []servicemanager.UIData, table *widgets.Table) {
	//oldRows := table.Rows
	table.Rows = [][]string{{"Name", "Balance", "Status"}}
	sort.Slice(serviceData, func(i, j int) bool {
		return serviceData[i].Name < serviceData[j].Name
	})
	for index, service := range serviceData {
		var oldServiceBalances []string
		const sep = ", "
		if oldUIData != nil {
			oldServiceBalances = strings.Split(oldUIData[index].Balance, sep)
		}
		var newServiceBalances []string
		for accountIndex, account := range service.Accounts {
			if oldServiceBalances != nil {
				if account.Balance
			} else {
				newServiceBalances = append(newServiceBalances, account.Balance.String())
			}
		}

		table.Rows = append(table.Rows, []string{service.Name, strings.Join(newServiceBalances, sep),
			service.Status})
	}
}
