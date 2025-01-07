package genesis

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	comettypes "github.com/cometbft/cometbft/types"
	cosmossdktypes "github.com/cosmos/cosmos-sdk/types"
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
	FeeMarket     *FeeMarket    `json:"feemarket"`
	Staking       *Staking      `json:"staking"`
}

type Staking struct {
	Params *StakingParams `json:"params"`
}

type StakingParams struct {
	BondDenom         string `json:"bond_denom"`
	HistoricalEntries uint64 `json:"historical_entries"`
	MaxEntries        uint32 `json:"max_entries"`
	MaxValidators     uint32 `json:"max_validators"`
	MinCommissionRate string `json:"min_commission_rate"`
	UnbondingTime     string `json:"unbonding_time"`
}

type FeeMarket struct {
	Params *FeeMarketParams `json:"params"`
}

type FeeMarketParams struct {
	BaseFee                  string `json:"base_fee"`
	BaseFeeChangeDenominator int    `json:"base_fee_change_denominator"`
	ElasticityMultiplier     int    `json:"elasticity_multiplier"`
	EnableHeight             string `json:"enable_height"`
	MinGasMultiplier         string `json:"min_gas_multiplier"`
	MinGasPrice              string `json:"min_gas_price"`
	NoBaseFee                bool   `json:"no_base_fee"`
}

type Bank struct {
	Supply []Denom `json:"supply"`
}

type RollappParams struct {
	Params struct {
		Da           string                  `json:"da"`
		DrsVersion   int                     `json:"drs_version"`
		MinGasPrices cosmossdktypes.DecCoins `json:"min_gas_prices"`
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

// GetGenesisAppState function retrieves the genesis file content using comet's
// native function, the problem here is that it takes time due to json.rawMessage
// for the app state itself
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

// GetAppStateFromGenesisFile function is a more minimalistic version of
// GetGenesisAppState, it only retrieves the relevant genesis information
// by unmarshalling bytes into the custom struct
func GetAppStateFromGenesisFile(home string) (*AppState, error) {
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

	return &as, err
}

func VerifyGenesisChainID(genesisPath, raID string) error {
	genesis, err := comettypes.GenesisDocFromFile(genesisPath)
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
	data, err := os.ReadFile(filepath)
	if err != nil {
		return "", fmt.Errorf("error opening file: %v", err)
	}

	var jsonObject map[string]interface{}
	err = json.Unmarshal(data, &jsonObject)
	if err != nil {
		return "", err
	}

	keys := make([]string, 0, len(jsonObject))
	for k := range jsonObject {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	sortedJSON, err := json.Marshal(jsonObject)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(sortedJSON)
	return hex.EncodeToString(hash[:]), nil
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

type ValidateGenesisRequest struct {
	RollappID             string `json:"rollapp-id"`
	SettlementChainID     string `json:"settlement-chain-id"`
	SettlementNodeAddress string `json:"settlement-node-address"`
}

const (
	ValidateGenesisURL = "https://genesis-validator.rollapp.network/validate-genesis"
)

func ValidateGenesis(raCfg roller.RollappConfig, raID string, hd consts.HubData) error {
	req := ValidateGenesisRequest{
		RollappID:             raID,
		SettlementChainID:     hd.ID,
		SettlementNodeAddress: hd.RpcUrl,
	}

	isChecksumValid, err := CompareGenesisChecksum(
		raCfg.Home,
		raID,
		hd,
	)

	if !isChecksumValid {
		return errors.New("genesis checksum mismatch")
	}

	if err != nil {
		return err
	}

	b, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := http.Post(ValidateGenesisURL, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	rb, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error validating genesis: %s", string(rb))
	}

	return nil
}
