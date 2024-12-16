package genesis

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cometbft/cometbft/types"
	comettypes "github.com/cometbft/cometbft/types"
	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config"
	"github.com/dymensionxyz/roller/utils/config/jsonconfig"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/rollapp"
	"github.com/dymensionxyz/roller/utils/roller"
	"github.com/dymensionxyz/roller/utils/sequencer"
)

type AppState struct {
	Bank          Bank          `json:"bank"`
	RollappParams RollappParams `json:"rollappparams"`
}

type Bank struct {
	Supply []Denom `json:"supply"`
}

type RollappParams struct {
	Params struct {
		Da         string `json:"da"`
		DrsVersion int    `json:"drs_version"`
	} `json:"params"`
}

type Denom struct {
	Denom string `json:"denom"`
}

func DownloadGenesis(home, genesisUrl string) error {
	genesisPath := GetGenesisFilePath(home)
	if genesisUrl == "" {
		return fmt.Errorf("RollApp's genesis url field is empty, contact the rollapp owner")
	}

	err := filesystem.DownloadFile(genesisUrl, genesisPath)
	if err != nil {
		return err
	}

	return nil
}

func GetGenesisAppState(home string) (*AppState, error) {
	genesis, err := comettypes.GenesisDocFromFile(GetGenesisFilePath(home))
	if err != nil {
		return nil, err
	}

	var as AppState
	j, _ := genesis.AppState.MarshalJSON()
	err = json.Unmarshal(j, &as)
	if err != nil {
		return nil, err
	}

	return &as, err
}

func GetDrsVersionFromGenesis(home string) (*AppState, error) {
	genesisFile, err := os.Open(GetGenesisFilePath(home))
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	// nolint:errcheck
	defer genesisFile.Close()

	var gs struct {
		AppState AppState `json:"app_state"`
	}
	err = json.NewDecoder(genesisFile).Decode(&gs)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling genesis file: %v", err)
	}

	as := gs.AppState

	j, err := json.MarshalIndent(as, "", "  ")
	fmt.Println(string(j))

	return &as, err
}

func VerifyGenesisChainID(genesisPath, raID string) error {
	genesis, err := types.GenesisDocFromFile(genesisPath)
	if err != nil {
		return err
	}

	if genesis.ChainID != raID {
		err := fmt.Errorf(
			"the genesis file ChainID (%s) does not match  the rollapp ID you're trying to initialize ("+
				"%s)",
			genesis.ChainID,
			raID,
		)
		return err
	}

	return nil
}

func calculateSHA256(filepath string) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", fmt.Errorf("error opening file: %v", err)
	}
	// nolint:errcheck
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("error calculating hash: %v", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func getRollappGenesisHash(raID string, hd consts.HubData) (string, error) {
	raResponse, err := rollapp.GetMetadataFromChain(raID, hd)
	if err != nil {
		return "", err
	}

	return raResponse.Rollapp.GenesisInfo.GenesisChecksum, nil
}

func CompareGenesisChecksum(root, raID string, hd consts.HubData) (bool, error) {
	genesisPath := GetGenesisFilePath(root)
	downloadedGenesisHash, err := calculateSHA256(genesisPath)
	if err != nil {
		pterm.Error.Println("failed to calculate hash of genesis file: ", err)
		return false, err
	}

	raGenesisHash, _ := getRollappGenesisHash(raID, hd)
	if downloadedGenesisHash != raGenesisHash {
		err = fmt.Errorf(
			"the hash of the downloaded file (%s) does not match the one registered with the rollapp (%s)",
			downloadedGenesisHash,
			raGenesisHash,
		)
		return false, err
	}

	return true, nil
}

func CompareRollappArchiveChecksum(
	filepath string,
	si sequencer.SnapshotInfo,
) (bool, error) {
	downloadedGenesisHash, err := calculateSHA256(filepath)
	if err != nil {
		pterm.Error.Println("failed to calculate hash of genesis file: ", err)
		return false, err
	}
	onChainHash := si.Checksum
	if downloadedGenesisHash != onChainHash {
		err = fmt.Errorf(
			"the hash of the downloaded file (%s) does not match the one registered with the rollapp (%s)",
			downloadedGenesisHash,
			onChainHash,
		)
		return false, err
	}

	return true, nil
}

func GetGenesisFilePath(root string) string {
	return filepath.Join(
		rollapp.RollappConfigDir(root),
		"genesis.json",
	)
}

func InitializeRollappGenesis(initConfig roller.RollappConfig) error {
	// totalTokenSupply, success := new(big.Int).SetString(consts.DefaultTokenSupply, 10)
	// if !success {
	// 	return fmt.Errorf("invalid token supply")
	// }
	// totalTokenSupply = totalTokenSupply.Mul(
	// 	totalTokenSupply, new(big.Int).Exp(
	// 		big.NewInt(10),
	// 		new(big.Int).SetUint64(uint64(initConfig.Decimals)), nil,
	// 	),
	// )

	// relayerGenesisBalance := new(big.Int).Div(totalTokenSupply, big.NewInt(10))
	// sequencerGenesisBalance := new(big.Int).Sub(totalTokenSupply, relayerGenesisBalance)
	// sequencerBalanceStr := sequencerGenesisBalance.String() + initConfig.Denom
	// rollappConfigDirPath := filepath.Join(initConfig.Home, consts.ConfigDirName.Rollapp)

	// genesisSequencerAccountCmd := exec.Command(
	// 	initConfig.RollappBinary,
	// 	"add-genesis-account",
	// 	consts.KeysIds.RollappSequencer,
	// 	sequencerBalanceStr,
	// 	"--keyring-backend",
	// 	"test",
	// 	"--home",
	// 	rollappConfigDirPath,
	// )
	// _, err := bash.ExecCommandWithStdout(genesisSequencerAccountCmd)
	// if err != nil {
	// 	return err
	// }

	err := UpdateGenesisParams(
		initConfig.Home,
		&initConfig,
	)
	if err != nil {
		return err
	}

	return nil
}

func UpdateGenesisParams(home string, raCfg *roller.RollappConfig) error {
	params := getDefaultGenesisParams(raCfg)
	addGenAccountCmd := GetAddGenesisAccountCmd(
		consts.KeysIds.RollappSequencer,
		consts.DefaultTokenSupply,
		raCfg,
	)

	genesisFilePath := filepath.Join(home, consts.ConfigDirName.Rollapp, "config", "genesis.json")
	_, err := bash.ExecCommandWithStdout(addGenAccountCmd)
	if err != nil {
		return err
	}

	return jsonconfig.UpdateJSONParams(genesisFilePath, params)
}

func getDefaultGenesisParams(
	raCfg *roller.RollappConfig,
) []config.PathValue {
	var params []config.PathValue

	commonParams := []config.PathValue{
		{Path: "app_state.mint.params.mint_denom", Value: raCfg.BaseDenom},
		{Path: "app_state.staking.params.bond_denom", Value: raCfg.BaseDenom},
		{Path: "app_state.gov.deposit_params.min_deposit.0.denom", Value: raCfg.BaseDenom},
		{Path: "consensus_params.block.max_gas", Value: "40000000"},
		{Path: "app_state.distribution.params.base_proposer_reward", Value: "0.8"},
		{Path: "app_state.distribution.params.community_tax", Value: "0.00002"},
		{Path: "app_state.gov.voting_params.voting_period", Value: "300s"},
		{Path: "app_state.staking.params.unbonding_time", Value: "3628800s"},
		{
			Path:  "app_state.bank.denom_metadata",
			Value: getBankDenomMetadata(raCfg.BaseDenom, raCfg.Decimals),
		},
		{Path: "app_state.claims.params.claims_denom", Value: raCfg.BaseDenom},
	}

	if raCfg.RollappVMType == consts.WASM_ROLLAPP {
		wasmParams := []config.PathValue{}
		params = append(commonParams, wasmParams...)
	}

	if raCfg.RollappVMType == consts.EVM_ROLLAPP {
		evmParams := []config.PathValue{
			{Path: "app_state.evm.params.evm_denom", Value: raCfg.BaseDenom},
			{Path: "app_state.feemarket.params.no_base_fee", Value: true},
			{Path: "app_state.evm.params.extra_eips", Value: []string{"3855"}},
			{Path: "app_state.feemarket.params.min_gas_price", Value: "0.0"},
		}
		params = append(commonParams, evmParams...)
	}

	return params
}

func getBankDenomMetadata(denom string, decimals uint) []BankDenomMetadata {
	displayDenom := denom[1:]

	metadata := []BankDenomMetadata{
		{
			Base: denom,
			DenomUnits: []BankDenomUnitMetadata{
				{
					Aliases:  []string{},
					Denom:    denom,
					Exponent: 0,
				},
				{
					Aliases:  []string{},
					Denom:    displayDenom,
					Exponent: decimals,
				},
			},
			Description: fmt.Sprintf("Denom metadata for %s (%s)", displayDenom, denom),
			Display:     displayDenom,
			Name:        displayDenom,
			Symbol:      strings.ToUpper(displayDenom),
		},
	}
	return metadata
}

type BankDenomMetadata struct {
	Base        string                  `json:"base"`
	DenomUnits  []BankDenomUnitMetadata `json:"denom_units"`
	Description string                  `json:"description"`
	Display     string                  `json:"display"`
	Name        string                  `json:"name"`
	Symbol      string                  `json:"symbol"`
}

type BankDenomUnitMetadata struct {
	Aliases  []string `json:"aliases"`
	Denom    string   `json:"denom"`
	Exponent uint     `json:"exponent"`
}

func GetAddGenesisAccountCmd(addr, amount string, raCfg *roller.RollappConfig) *exec.Cmd {
	home := raCfg.Home
	cmd := exec.Command(
		consts.Executables.RollappEVM,
		"add-genesis-account",
		addr,
		fmt.Sprintf("%s%s", amount, raCfg.BaseDenom),
		"--home",
		fmt.Sprintf("%s/%s", home, consts.ConfigDirName.Rollapp),
		"--keyring-backend",
		string(raCfg.KeyringBackend),
	)

	return cmd
}
