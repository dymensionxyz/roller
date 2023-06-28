package initconfig

import (
	"fmt"
	"io/ioutil"
	"math/big"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"

	"os/exec"
	"path/filepath"

	"github.com/tidwall/sjson"
)

func initializeRollappGenesis(initConfig utils.RollappConfig) error {
	totalTokenSupply, success := new(big.Int).SetString(initConfig.TokenSupply, 10)
	if !success {
		return fmt.Errorf("invalid token supply")
	}
	relayerGenesisBalance := new(big.Int).Div(totalTokenSupply, big.NewInt(10))
	sequencerGenesisBalance := new(big.Int).Sub(totalTokenSupply, relayerGenesisBalance)
	sequencerBalanceStr := sequencerGenesisBalance.String() + initConfig.Denom
	relayerBalanceStr := relayerGenesisBalance.String() + initConfig.Denom
	rollappConfigDirPath := filepath.Join(initConfig.Home, consts.ConfigDirName.Rollapp)
	genesisSequencerAccountCmd := exec.Command(initConfig.RollappBinary, "add-genesis-account",
		consts.KeyNames.RollappSequencer, sequencerBalanceStr, "--keyring-backend", "test", "--home", rollappConfigDirPath)
	_, err := utils.ExecBashCommand(genesisSequencerAccountCmd)
	if err != nil {
		return err
	}
	rlyRollappAddress, err := utils.GetRelayerAddress(initConfig.Home, initConfig.RollappID)
	if err != nil {
		return err
	}
	genesisRelayerAccountCmd := exec.Command(initConfig.RollappBinary, "add-genesis-account",
		rlyRollappAddress, relayerBalanceStr, "--home", rollappConfigDirPath)
	_, err = utils.ExecBashCommand(genesisRelayerAccountCmd)
	if err != nil {
		return err
	}
	err = updateGenesisParams(GetGenesisFilePath(initConfig.Home), initConfig.Denom)
	if err != nil {
		return err
	}
	return nil
}

func GetGenesisFilePath(root string) string {
	return filepath.Join(RollappConfigDir(root), "genesis.json")
}

type PathValue struct {
	Path  string
	Value interface{}
}

// TODO(#130): fix to support epochs
func getDefaultGenesisParams(denom string) []PathValue {
	return []PathValue{
		{"app_state.mint.params.mint_denom", denom},
		{"app_state.staking.params.bond_denom", denom},
		{"app_state.crisis.constant_fee.denom", denom},
		{"app_state.evm.params.evm_denom", denom},
		{"app_state.gov.deposit_params.min_deposit.0.denom", denom},
		{"consensus_params.block.max_gas", "40000000"},
		{"app_state.feemarket.params.no_base_fee", true},
		{"app_state.distribution.params.base_proposer_reward", "0.8"},
		{"app_state.distribution.params.community_tax", "0.00002"},
		{"app_state.gov.voting_params.voting_period", "300s"},
		{"app_state.staking.params.unbonding_time", "3628800s"},
	}
}

func UpdateJSONParams(jsonFilePath string, params []PathValue) error {
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
	return UpdateJSONParams(genesisFilePath, params)
}
