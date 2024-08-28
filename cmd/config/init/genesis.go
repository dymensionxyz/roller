package initconfig

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tidwall/sjson"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config"
)

// const (
// 	totalSupplyToStakingRatio = 2
// )

type PathValue struct {
	Path  string
	Value interface{}
}

// TODO(#130): fix to support epochs
func getDefaultGenesisParams(
	sequencerAddr string, raCfg *config.RollappConfig,
) []PathValue {
	return []PathValue{
		// these should be injected from the genesis creator
		{"consensus_params.block.max_gas", "40000000"},
		{"app_state.feemarket.params.no_base_fee", true},
		{"app_state.feemarket.params.min_gas_price", "0.0"},
		{"app_state.distribution.params.base_proposer_reward", "0.8"},
		{"app_state.distribution.params.community_tax", "0.00002"},
		{"app_state.gov.voting_params.voting_period", "300s"},
		{"app_state.bank.denom_metadata", getBankDenomMetadata(raCfg.BaseDenom, raCfg.Decimals)},
		{"app_state.denommetadata.params.allowed_addresses.0", sequencerAddr},
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

	// nolint:gofumpt
	err = os.WriteFile(jsonFilePath, []byte(jsonFileContentString), 0o644)
	if err != nil {
		return err
	}
	return nil
}

func UpdateGenesisParams(home string, raCfg *config.RollappConfig) error {
	sa, err := GetRollappSequencerAddress(home)
	if err != nil {
		return err
	}
	params := getDefaultGenesisParams(sa, raCfg)
	addGenAccountCmd := GetAddGenesisAccountCmd(
		consts.KeysIds.RollappSequencer,
		consts.DefaultTokenSupply,
		raCfg,
	)

	_, err = bash.ExecCommandWithStdout(addGenAccountCmd)
	if err != nil {
		return err
	}

	genesisFilePath := filepath.Join(home, consts.ConfigDirName.Rollapp, "config", "genesis.json")
	return UpdateJSONParams(genesisFilePath, params)
}

func GetAddGenesisAccountCmd(addr, amount string, raCfg *config.RollappConfig) *exec.Cmd {
	home := raCfg.Home
	cmd := exec.Command(
		consts.Executables.RollappEVM,
		"add-genesis-account",
		addr,
		fmt.Sprintf("%s%s", amount, raCfg.BaseDenom),
		"--home",
		fmt.Sprintf("%s/%s", home, consts.ConfigDirName.Rollapp),
		"--keyring-backend",
		"test",
	)

	return cmd
}

func getGenesisOperatorAddress(home string) (string, error) {
	rollappConfigDirPath := filepath.Join(home, consts.ConfigDirName.Rollapp)
	getOperatorAddrCommand := exec.Command(
		consts.Executables.RollappEVM,
		"keys",
		"show",
		consts.KeysIds.RollappSequencer,
		"-a",
		"--keyring-backend",
		"test",
		"--home",
		rollappConfigDirPath,
		"--bech",
		"val",
	)

	addr, err := bash.ExecCommandWithStdout(getOperatorAddrCommand)
	if err != nil {
		fmt.Println("val addr failed")
		return "", err
	}

	a := strings.TrimSpace(addr.String())
	return a, nil
}

func GetRollappSequencerAddress(home string) (string, error) {
	seqKeyConfig := utils.KeyConfig{
		Dir:         consts.ConfigDirName.Rollapp,
		ID:          consts.KeysIds.RollappSequencer,
		ChainBinary: consts.Executables.RollappEVM,
		Type:        consts.EVM_ROLLAPP,
	}
	addr, err := utils.GetAddressBinary(seqKeyConfig, home)
	if err != nil {
		return "", err
	}

	return addr, nil
}

// func generateGenesisTx(initConfig config.RollappConfig) error {
// 	err := registerSequencerAsGoverner(initConfig)
// 	if err != nil {
// 		return fmt.Errorf("failed to execute gentx command: %v", err)
// 	}
// 	// collect gentx
// 	rollappConfigDirPath := filepath.Join(initConfig.Home, consts.ConfigDirName.Rollapp)
// 	collectGentx := exec.Command(
// 		initConfig.RollappBinary,
// 		"collect-gentxs",
// 		"--home",
// 		rollappConfigDirPath,
// 	)
// 	_, err = utils.ExecCommandWithStdout(collectGentx)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }
//
// // registerSequencerAsGoverner registers the sequencer as a governor of the rollapp chain.
// // currently it sets the staking amount to half of the total token supply.
// // TODO: make the staking amount configurable
// func registerSequencerAsGoverner(initConfig config.RollappConfig) error {
// 	totalSupply, err := strconv.Atoi(consts.DefaultTokenSupply)
// 	if err != nil {
// 		return fmt.Errorf("error converting string to integer: %w", err)
// 	}
//
// 	// Convert to token supply with decimals
// 	stakedSupply := big.NewInt(int64(totalSupply / totalSupplyToStakingRatio))
// 	multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(initConfig.Decimals)), nil)
// 	stakedSupply.Mul(stakedSupply, multiplier)
//
// 	// Build and run the gentx command
// 	rollappConfigDirPath := filepath.Join(initConfig.Home, consts.ConfigDirName.Rollapp)
// 	gentxCmd := exec.Command(
// 		initConfig.RollappBinary,
// 		"gentx",
// 		consts.KeysIds.RollappSequencer,
// 		fmt.Sprint(
// 			stakedSupply,
// 			initConfig.Denom,
// 		),
// 		"--chain-id",
// 		initConfig.RollappID,
// 		"--keyring-backend",
// 		"test",
// 		"--home",
// 		rollappConfigDirPath,
// 	)
// 	_, err = utils.ExecCommandWithStdout(gentxCmd)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }
