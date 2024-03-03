package initconfig

import (
	"fmt"
	"math/big"
	"os"
	"strconv"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"

	"os/exec"
	"path/filepath"

	"github.com/tidwall/sjson"
)

const (
	totalSupplyToStakingRatio = 2
)

func initializeRollappGenesis(initConfig config.RollappConfig) error {
	totalTokenSupply, success := new(big.Int).SetString(initConfig.TokenSupply, 10)
	if !success {
		return fmt.Errorf("invalid token supply")
	}
	totalTokenSupply = totalTokenSupply.Mul(totalTokenSupply, new(big.Int).Exp(big.NewInt(10),
		new(big.Int).SetUint64(uint64(initConfig.Decimals)), nil))
	relayerGenesisBalance := new(big.Int).Div(totalTokenSupply, big.NewInt(10))
	sequencerGenesisBalance := new(big.Int).Sub(totalTokenSupply, relayerGenesisBalance)
	sequencerBalanceStr := sequencerGenesisBalance.String() + initConfig.Denom
	relayerBalanceStr := relayerGenesisBalance.String() + initConfig.Denom
	rollappConfigDirPath := filepath.Join(initConfig.Home, consts.ConfigDirName.Rollapp)
	genesisSequencerAccountCmd := exec.Command(initConfig.RollappBinary, "add-genesis-account",
		consts.KeysIds.RollappSequencer, sequencerBalanceStr, "--keyring-backend", "test", "--home", rollappConfigDirPath)
	_, err := utils.ExecBashCommandWithStdout(genesisSequencerAccountCmd)
	if err != nil {
		return err
	}
	rlyRollappAddress, err := utils.GetRelayerAddress(initConfig.Home, initConfig.RollappID)
	if err != nil {
		return err
	}
	genesisRelayerAccountCmd := exec.Command(initConfig.RollappBinary, "add-genesis-account",
		rlyRollappAddress, relayerBalanceStr, "--keyring-backend", "test", "--home", rollappConfigDirPath)
	_, err = utils.ExecBashCommandWithStdout(genesisRelayerAccountCmd)
	if err != nil {
		return err
	}
	err = generateGenesisTx(initConfig)
	if err != nil {
		return err
	}
	err = updateGenesisParams(GetGenesisFilePath(initConfig.Home), initConfig.Denom, initConfig.Decimals)
	if err != nil {
		return err
	}

	err = createTokenMetadaJSON(filepath.Join(RollappConfigDir(initConfig.Home), "tokenmetadata.json"), initConfig.Denom, initConfig.Decimals)
	if err != nil {
		return err
	}

	return nil
}

func GetGenesisFilePath(root string) string {
	return filepath.Join(RollappConfigDir(root),
		"genesis.json")
}

type PathValue struct {
	Path  string
	Value interface{}
}

// TODO(#130): fix to support epochs
func getDefaultGenesisParams(denom string, decimals uint) []PathValue {
	return []PathValue{
		{"app_state.mint.params.mint_denom", denom},
		{"app_state.staking.params.bond_denom", denom},
		{"app_state.crisis.constant_fee.denom", denom},
		{"app_state.evm.params.evm_denom", denom},
		{"app_state.gov.deposit_params.min_deposit.0.denom", denom},
		{"consensus_params.block.max_gas", "40000000"},
		{"app_state.feemarket.params.no_base_fee", true},
		{"app_state.feemarket.params.min_gas_price", "0.0"},
		{"app_state.distribution.params.base_proposer_reward", "0.8"},
		{"app_state.distribution.params.community_tax", "0.00002"},
		{"app_state.gov.voting_params.voting_period", "300s"},
		{"app_state.staking.params.unbonding_time", "3628800s"},
		{"app_state.bank.denom_metadata", getBankDenomMetadata(denom, decimals)},
	}
}

func UpdateJSONParams(jsonFilePath string, params []PathValue) error {
	jsonFileContent, err := os.ReadFile(jsonFilePath)
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
	err = os.WriteFile(jsonFilePath, []byte(jsonFileContentString), 0644)
	if err != nil {
		return err
	}
	return nil
}

func updateGenesisParams(genesisFilePath string, denom string, decimals uint) error {
	params := getDefaultGenesisParams(denom, decimals)
	return UpdateJSONParams(genesisFilePath, params)
}

func generateGenesisTx(initConfig config.RollappConfig) error {
	err := registerSequencerAsGoverner(initConfig)
	if err != nil {
		return err
	}
	// collect gentx
	rollappConfigDirPath := filepath.Join(initConfig.Home, consts.ConfigDirName.Rollapp)
	collectGentx := exec.Command(initConfig.RollappBinary, "collect-gentxs", "--home", rollappConfigDirPath)
	_, err = utils.ExecBashCommandWithStdout(collectGentx)
	if err != nil {
		return err
	}
	return nil

}

// registerSequencerAsGoverner registers the sequencer as a governor of the rollapp chain.
// currently it sets the staking amount to half of the total token supply.
// TODO: make the staking amount configurable
func registerSequencerAsGoverner(initConfig config.RollappConfig) error {
	totalSupply, err := strconv.Atoi(initConfig.TokenSupply)
	if err != nil {
		return fmt.Errorf("error converting string to integer: %w", err)
	}
	// Convert to token supply with decimals
	stakedSupply := big.NewInt(int64(totalSupply / totalSupplyToStakingRatio))
	multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(initConfig.Decimals)), nil)
	stakedSupply.Mul(stakedSupply, multiplier)
	// Build and run the gentx command
	rollappConfigDirPath := filepath.Join(initConfig.Home, consts.ConfigDirName.Rollapp)
	gentxCmd := exec.Command(initConfig.RollappBinary, "gentx", consts.KeysIds.RollappSequencer,
		fmt.Sprint(stakedSupply, initConfig.Denom), "--chain-id", initConfig.RollappID, "--keyring-backend", "test", "--home", rollappConfigDirPath)
	_, err = utils.ExecBashCommandWithStdout(gentxCmd)
	if err != nil {
		return err
	}
	return nil
}
