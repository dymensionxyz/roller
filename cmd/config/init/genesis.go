package initconfig

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

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
