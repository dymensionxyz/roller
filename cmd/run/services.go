package run

import (
	"github.com/dymensionxyz/roller/cmd/utils"
	"log"
)

type ServiceData struct {
	Name    string
	Balance string
	Status  string
}

type fetchResult struct {
	data *utils.AccountData
	err  error
	id   int
}

func fetchServicesData(rollappConfig utils.RollappConfig, logger *log.Logger) ([]ServiceData, error) {
	fetchFuncs := []func(utils.RollappConfig) (*utils.AccountData, error){
		utils.GetSequencerData,
		utils.GetHubRlyAccData,
		utils.GetRolRlyAccData,
		utils.GetCelLCAccData,
	}
	results := fetchAsync(fetchFuncs, rollappConfig)
	data := processDataResults(results, len(fetchFuncs), logger)
	return buildServiceData(data, rollappConfig), nil
}

func getInitialServiceData() []ServiceData {
	return []ServiceData{
		{
			Name:    "Sequencer",
			Balance: "Fetching...",
			Status:  "Active",
		},
		{
			Name:    "DA Light Client",
			Balance: "Fetching...",
			Status:  "Active",
		},
		{
			Name:    "Relayer",
			Balance: "Fetching...",
			Status:  "Starting...",
		},
	}
}

func fetchAsync(fetchFuncs []func(utils.RollappConfig) (*utils.AccountData, error), rollappConfig utils.RollappConfig) chan fetchResult {
	results := make(chan fetchResult, len(fetchFuncs))
	for i, fn := range fetchFuncs {
		go func(id int, fn func(utils.RollappConfig) (*utils.AccountData, error)) {
			data, err := fn(rollappConfig)
			results <- fetchResult{data, err, id}
		}(i, fn)
	}
	return results
}
