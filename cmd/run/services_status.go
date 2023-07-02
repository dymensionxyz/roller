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

func activeIfSufficientBalance(currentBalance, threshold *big.Int) string {
	if currentBalance.Cmp(threshold) >= 0 {
		return "Active"
	} else {
		return "Stopped"
	}
}

func buildServiceData(data []*utils.AccountData, rollappConfig utils.RollappConfig) []ServiceData {
	rolRlyData := data[2]
	return []ServiceData{
		{
			Name:    "Sequencer",
			Balance: data[0].Balance.String() + consts.Denoms.Hub,
			// TODO: for now, we just check if the balance of the rollapp relayer is greater than 0
			// in the future, we should have a better way to check the rollapp health.
			Status: activeIfSufficientBalance(rolRlyData.Balance, big.NewInt(1)),
		},
		{
			Name:    "DA Light Client",
			Balance: data[3].Balance.String() + consts.Denoms.Celestia,
			Status:  activeIfSufficientBalance(data[3].Balance, consts.OneDAWritePrice),
		},
		{
			Name: "Relayer",
			Balance: data[1].Balance.String() + consts.Denoms.Hub + ", " +
				rolRlyData.Balance.String() + rollappConfig.Denom,
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
	servicesStatusTable := buildUITable(termWidth)
	servicesInfoTable := getServicesInfo(rollappConfig, termWidth)
	serviceData := getInitialServiceData()

	updateUITable(serviceData, servicesStatusTable)
	ui.Render(p, servicesStatusTable, servicesInfoTable)

	events := ui.PollEvents()
	ticker := time.NewTicker(time.Second * 1).C

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
			} else {
				config.logger.Printf("Fetched services data successfully %s", serviceData)
			}
			updateUITable(serviceData, config.table)
			ui.Render(config.table)
		}
	}
}

type ServiceStatusConfig struct {
	rollappConfig utils.RollappConfig
	logger        *log.Logger
	table         *widgets.Table
}
