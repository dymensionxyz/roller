package run

import (
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"log"
	"math/big"
	"path/filepath"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

func processDataResults(results chan fetchResult, size int, logger *log.Logger) []*utils.AccountData {
	data := make([]*utils.AccountData, size)
	for i := 0; i < size; i++ {
		res := <-results
		if res.err != nil {
			logger.Println(res.err)
			data[res.id] = &utils.AccountData{
				Address: "",
				Balance: big.NewInt(0),
			}
		} else {
			data[res.id] = res.data
		}
	}
	return data
}

func buildServiceData(data []*utils.AccountData, rollappConfig utils.RollappConfig) []ServiceData {
	return []ServiceData{
		{
			Name:    "Sequencer",
			Balance: data[0].Balance.String() + consts.Denoms.Hub,
			Status:  "Active",
		},
		{
			Name:    "DA Light Client",
			Balance: data[3].Balance.String() + consts.Denoms.Celestia,
			Status:  "Active",
		},
		{
			Name: "Relayer",
			Balance: data[1].Balance.String() + consts.Denoms.Hub + ", " +
				data[2].Balance.String() + rollappConfig.Denom,
			Status: "Starting...",
		},
	}
}

func PrintServicesStatus(rollappConfig utils.RollappConfig) {
	logger := utils.GetLogger(filepath.Join(rollappConfig.Home, "roller.log"))
	initializeUI()
	defer ui.Close()

	termWidth, _ := ui.TerminalDimensions()

	p := buildUIParagraph(termWidth)
	table := buildUITable(termWidth)
	serviceData := getInitialServiceData()

	updateUITable(serviceData, table)
	ui.Render(p, table)

	events := ui.PollEvents()
	ticker := time.NewTicker(time.Second * 5).C

	config := ServiceStatusConfig{
		rollappConfig: rollappConfig,
		logger:        logger,
		table:         table,
		p:             p,
	}

	eventLoop(events, ticker, config)
}

func eventLoop(events <-chan ui.Event, ticker <-chan time.Time, config ServiceStatusConfig) {
	for {
		select {
		case e := <-events:
			if e.ID == "q" || e.ID == "<C-c>" {
				return
			}
		case <-ticker:
			config.logger.Println("Fetching service data...")
			serviceData, err := fetchServicesData(config.rollappConfig, config.logger)
			if err != nil {
				config.logger.Printf("Error: failed to fetch service data: %v", err)
			} else {
				config.logger.Printf("Fetched services data successfully %s", serviceData)

			}
			updateUITable(serviceData, config.table)
			ui.Render(config.p, config.table)
		}
	}
}

type ServiceStatusConfig struct {
	rollappConfig utils.RollappConfig
	logger        *log.Logger
	table         *widgets.Table
	p             *widgets.Paragraph
}
