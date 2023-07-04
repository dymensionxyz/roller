package run

import (
	"log"
	"math/big"
	"path/filepath"
	"time"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	datalayer "github.com/dymensionxyz/roller/data_layer"

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

func activeIfSufficientBalance(currentBalance, threshold *big.Int) string {
	if currentBalance.Cmp(threshold) >= 0 {
		return "Active"
	} else {
		return "Stopped"
	}
}

func buildServiceData(data []*utils.AccountData, rollappConfig config.RollappConfig) []ServiceData {
	damanager := datalayer.NewDAManager(rollappConfig.DA, rollappConfig.Home)

	var servicedata = []ServiceData{
		{
			Name:    "Sequencer",
			Balance: data[0].Balance.String() + consts.Denoms.Hub,
			Status:  activeIfSufficientBalance(data[0].Balance, big.NewInt(1)),
		},
		{
			Name: "Relayer",
			Balance: data[1].Balance.String() + consts.Denoms.Hub + ", " +
				data[1].Balance.String() + rollappConfig.Denom,
			Status: "Starting...",
		},
	}

	if damanager.GetLightNodeEndpoint() != "" {
		servicedata = append(servicedata, ServiceData{
			Name:    "DA Light Client",
			Balance: data[3].Balance.String() + consts.Denoms.Celestia,
			Status:  activeIfSufficientBalance(data[3].Balance, consts.OneDAWritePrice),
		})
	}
	return servicedata
}

func RenderUI(rollappConfig config.RollappConfig) {
	logger := utils.GetLogger(filepath.Join(rollappConfig.Home, "roller.log"))
	initializeUI()
	defer ui.Close()

	termWidth, _ := ui.TerminalDimensions()

	p := buildTitleParagraph(rollappConfig, termWidth)
	servicesStatusTable := NewServiceStatusTable(termWidth)
	servicesInfoTable := NewServicesInfoTable(rollappConfig, termWidth)

	serviceData, err := fetchServicesData(rollappConfig, logger)
	if err != nil {
		logger.Printf("Error: failed to fetch service data: %v", err)
		serviceData = []ServiceData{}
	}
	updateUITable(serviceData, servicesStatusTable)
	ui.Render(p, servicesStatusTable, servicesInfoTable)

	events := ui.PollEvents()
	ticker := time.NewTicker(time.Second * 5).C

	//TODO: the renderer should be a struct that holds the config and the tables
	config := ServiceStatusConfig{
		rollappConfig: rollappConfig,
		logger:        logger,
		table:         servicesStatusTable,
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
				serviceData = []ServiceData{}
			} else {
				config.logger.Printf("Fetched services data successfully %s", serviceData)
			}
			updateUITable(serviceData, config.table)
			ui.Render(config.table)
		}
	}
}

type ServiceStatusConfig struct {
	rollappConfig config.RollappConfig
	logger        *log.Logger
	table         *widgets.Table
}
