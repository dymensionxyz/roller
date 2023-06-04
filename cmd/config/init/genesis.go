package init

import (
	"io/ioutil"
	"os"

	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/tidwall/sjson"
)

func initializeRollappGenesis(rollappExecutablePath string, decimals uint64, denom string) error {
	zeros := decimals + 9
	tokenAmount := "1" + fmt.Sprintf("%0*d", zeros, 0) + denom
	rollappConfigDirPath := filepath.Join(os.Getenv("HOME"), configDirName.Rollapp)
	genesisSequencerAccountCmd := exec.Command(rollappExecutablePath, "add-genesis-account", keyNames.RollappSequencer, tokenAmount, "--keyring-backend", "test", "--home", rollappConfigDirPath)
	err := genesisSequencerAccountCmd.Run()
	if err != nil {
		return err
	}
	err = updateGenesisParams(filepath.Join(rollappConfigDirPath, "config", "genesis.json"), denom)
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

func updateGenesisParams(genesisFilePath string, denom string) error {
	genesisFileContent, err := ioutil.ReadFile(genesisFilePath)
	if err != nil {
		return err
	}
	genesisFileContentString := string(genesisFileContent)
	for _, param := range getDefaultGenesisParams(denom) {
		genesisFileContentString, err = sjson.Set(genesisFileContentString, param.Path, param.Value)
		if err != nil {
			return err
		}
	}
	err = ioutil.WriteFile(genesisFilePath, []byte(genesisFileContentString), 0644)
	if err != nil {
		return err
	}
	return nil
}
