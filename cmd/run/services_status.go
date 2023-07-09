package run

import (
	"log"
	"path/filepath"
	"time"

	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	servicemanager "github.com/dymensionxyz/roller/utils/service_manager"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

func RenderUI(rollappConfig config.RollappConfig, manager *servicemanager.ServiceConfig) {
	logger := utils.GetLogger(filepath.Join(rollappConfig.Home, "roller.log"))
	initializeUI()
	defer ui.Close()

	termWidth, _ := ui.TerminalDimensions()

	p := buildTitleParagraph(rollappConfig, termWidth)
	servicesStatusTable := NewServiceStatusTable(termWidth)
	servicesInfoTable := NewServicesInfoTable(rollappConfig, termWidth)

	manager.InitServicesData(rollappConfig)
	updateUITable(manager.GetUIData(), servicesStatusTable)
	ui.Render(p, servicesStatusTable, servicesInfoTable)

	//TODO: the renderer should be a struct that holds the config and the tables
	config := ServiceStatusConfig{
		rollappConfig: rollappConfig,
		logger:        logger,
		table:         servicesStatusTable,
	}
	events := ui.PollEvents()
	ticker := time.NewTicker(time.Second * 5).C

	eventLoop(events, ticker, manager, config)
}

func eventLoop(events <-chan ui.Event, ticker <-chan time.Time, manager *servicemanager.ServiceConfig,
	config ServiceStatusConfig) {
	for {
		select {
		case e := <-events:
			if e.ID == "q" || e.ID == "<C-c>" {
				return
			}
		case <-ticker:
			manager.Logger.Println("Fetching service data...")
			manager.FetchServicesData(config.rollappConfig)
			updateUITable(manager.GetUIData(), config.table)
			ui.Render(config.table)
		}
	}
}

type ServiceStatusConfig struct {
	rollappConfig config.RollappConfig
	logger        *log.Logger
	table         *widgets.Table
}
