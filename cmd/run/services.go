package run

import (
	"log"

	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	datalayer "github.com/dymensionxyz/roller/data_layer"
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

func fetchServicesData(rollappConfig config.RollappConfig, logger *log.Logger) ([]ServiceData, error) {
	damanager := datalayer.NewDAManager(rollappConfig.DA, rollappConfig.Home)

	//TODO: avoid requiring passing rollappConfig to every function
	fetchFuncs := []func(config.RollappConfig) (*utils.AccountData, error){
		utils.GetSequencerData,
		utils.GetHubRlyAccData,
		utils.GetRolRlyAccData,
		damanager.GetDAAccData,
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

func fetchAsync(fetchFuncs []func(config.RollappConfig) (*utils.AccountData, error), rollappConfig config.RollappConfig) chan fetchResult {
	results := make(chan fetchResult, len(fetchFuncs))
	for i, fn := range fetchFuncs {
		go func(id int, fn func(config.RollappConfig) (*utils.AccountData, error)) {
			data, err := fn(rollappConfig)
			results <- fetchResult{data, err, id}
		}(i, fn)
	}
	return results
}
