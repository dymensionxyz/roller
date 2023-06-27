package run

import (
	"github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"log"
)

func initializeUI() {
	if err := termui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
}

func buildUIParagraph(termWidth int) *widgets.Paragraph {
	p := widgets.NewParagraph()
	p.Text = "💈 The rollapp is running successfully on your local machine!"
	p.SetRect(0, 0, termWidth, 3)
	return p
}

func buildUITable(termWidth int) *widgets.Table {
	table := widgets.NewTable()
	table.RowStyles[0] = termui.NewStyle(termui.ColorWhite, termui.ColorClear, termui.ModifierBold)
	table.SetRect(0, 3, termWidth, 12)
	table.Title = "Services Status"
	return table
}

func updateUITable(serviceData []ServiceData, table *widgets.Table) {
	table.Rows = append(table.Rows, []string{"Name", "Balance", "Status"})
	for _, data := range serviceData {
		table.Rows = append(table.Rows, []string{data.Name, data.Balance, data.Status})
	}
}
