package init

import (
	"io/ioutil"

	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/tidwall/sjson"
)

func initializeRollappGenesis(initConfig InitConfig) error {
	zeros := initConfig.Decimals + 9
	tokenAmount := "1" + fmt.Sprintf("%0*d", zeros, 0) + initConfig.Denom
	rollappConfigDirPath := filepath.Join(initConfig.Home, configDirName.Rollapp)
	genesisSequencerAccountCmd := exec.Command(initConfig.RollappBinary, "add-genesis-account", keyNames.RollappSequencer, tokenAmount, "--keyring-backend", "test", "--home", rollappConfigDirPath)
	err := genesisSequencerAccountCmd.Run()
	if err != nil {
		return err
	}
	err = updateGenesisParams(filepath.Join(rollappConfigDirPath, "config", "genesis.json"), initConfig.Denom)
	if err != nil {
		return err
	}
	return nil
}

type pathValue struct {
	Path  string
	Value interface{}
}

func getDefaultGenesisParams(denom string) []pathValue {
	return []pathValue{
		{"app_state.mint.params.mint_denom", denom},
		{"app_state.staking.params.bond_denom", denom},
		{"app_state.crisis.constant_fee.denom", denom},
		{"app_state.evm.params.evm_denom", denom},
		{"app_state.gov.deposit_params.min_deposit.0.denom", denom},
		{"consensus_params.block.max_gas", "40000000"},
		{"app_state.feemarket.params.no_base_fee", true},
		{"app_state.mint.params.blocks_per_year", "157680000"},
		{"app_state.distribution.params.base_proposer_reward", "0.8"},
		{"app_state.distribution.params.community_tax", "0.00002"},
		{"app_state.gov.voting_params.voting_period", "300s"},
		{"app_state.staking.params.unbonding_time", "3628800s"},
	}
}

func updateJSONParams(jsonFilePath string, params []pathValue) error {
	jsonFileContent, err := ioutil.ReadFile(jsonFilePath)
	if err != nil {
		return err
	}
	jsonFileContentString := string(jsonFileContent)
	for _, param := range params {
		jsonFileContentString, err = sjson.Set(jsonFileContentString, param.Path, param.Value)
		if err != nil {
			return err
		}
	}
	err = ioutil.WriteFile(jsonFilePath, []byte(jsonFileContentString), 0644)
	if err != nil {
		return err
	}
	return nil
}

func updateGenesisParams(genesisFilePath string, denom string) error {
	params := getDefaultGenesisParams(denom)
	return updateJSONParams(genesisFilePath, params)
}
