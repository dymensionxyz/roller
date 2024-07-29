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
	"github.com/dymensionxyz/roller/config"
)

const (
	totalSupplyToStakingRatio = 2
)

type PathValue struct {
	Path  string
	Value interface{}
}

func GetGenesisFilePath(root string) string {
	return filepath.Join(RollappConfigDir(root),
		"genesis.json")
}

// TODO(#130): fix to support epochs
func getDefaultGenesisParams(
	sequencerAddr, genesisOperatorAddress string, raCfg *config.RollappConfig,
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

		{"app_state.sequencers.genesis_operator_address", genesisOperatorAddress},
		{
			"app_state.hubgenesis.params.genesis_triggerer_allowlist.0",
			map[string]string{"address": sequencerAddr},
		},
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
	err = os.WriteFile(jsonFilePath, []byte(jsonFileContentString), 0644)
	if err != nil {
		return err
	}
	return nil
}

func UpdateGenesisParams(home string, raCfg *config.RollappConfig) error {
	oa, err := getGenesisOperatorAddress(home)
	if err != nil {
		return err
	}
	cfg, err := config.LoadConfigFromTOML(home)
	if err != nil {
		return err
	}

	sa, err := getSequencerAddress(home)
	if err != nil {
		return err
	}
	params := getDefaultGenesisParams(sa, oa, raCfg)

	// TODO: move to generalized helper
	addGenAccountCmd := exec.Command(
		consts.Executables.RollappEVM,
		"add-genesis-account",
		sa,
		fmt.Sprintf("%s%s", consts.DefaultTokenSupply, cfg.BaseDenom),
		"--home",
		fmt.Sprintf("%s/%s", home, consts.ConfigDirName.Rollapp),
		"--keyring-backend",
		"test",
	)

	_, err = utils.ExecBashCommandWithStdout(addGenAccountCmd)
	if err != nil {
		return err
	}

	genesisFilePath := filepath.Join(home, consts.ConfigDirName.Rollapp, "config", "genesis.json")
	return UpdateJSONParams(genesisFilePath, params)
}

func GetAddGenesisAccountCmd(addr, amount string, raCfg *config.RollappConfig) *exec.Cmd {
	home := utils.GetRollerRootDir()
	cmd := exec.Command(
		consts.Executables.RollappEVM,
		"add-genesis-account",
		addr,
		fmt.Sprintf("%s%s", consts.DefaultTokenSupply, raCfg.BaseDenom),
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

	addr, err := utils.ExecBashCommandWithStdout(getOperatorAddrCommand)
	if err != nil {
		fmt.Println("val addr failed")
		return "", err
	}

	a := strings.TrimSpace(addr.String())
	return a, nil
}

func getSequencerAddress(home string) (string, error) {
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
	)

	addr, err := utils.ExecBashCommandWithStdout(getOperatorAddrCommand)
	if err != nil {
		fmt.Println("seq addr failed")
		return "", err
	}

	a := strings.TrimSpace(addr.String())
	return a, nil
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
// 	_, err = utils.ExecBashCommandWithStdout(collectGentx)
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
// 	_, err = utils.ExecBashCommandWithStdout(gentxCmd)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }
