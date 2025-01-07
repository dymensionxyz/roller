package rollapp

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	cosmossdktypes "github.com/cosmos/cosmos-sdk/types"
	dymensiontypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
	bashutils "github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/roller"
	"github.com/dymensionxyz/roller/version"
)

func GetHomeDir(home string) string {
	return filepath.Join(home, consts.ConfigDirName.Rollapp)
}

func GetCurrentHeight() (*BlockInformation, error) {
	cmd := getCurrentBlockCmd()
	out, err := bashutils.ExecCommandWithStdout(cmd)
	if err != nil {
		return nil, err
	}

	var blockInfo BlockInformation
	err = json.Unmarshal(out.Bytes(), &blockInfo)
	if err != nil {
		return nil, err
	}

	return &blockInfo, nil
}

func getCurrentBlockCmd() *exec.Cmd {
	cmd := exec.Command(
		consts.Executables.RollappEVM,
		"q",
		"block",
	)
	return cmd
}

func GetInitialSequencerAddress(raID string, hd consts.HubData) (string, error) {
	cmd := GetShowRollappCmd(raID, hd)
	out, err := bashutils.ExecCommandWithStdout(cmd)
	if err != nil {
		fmt.Println(err)
	}

	var ra dymensiontypes.QueryGetRollappResponse
	_ = json.Unmarshal(out.Bytes(), &ra)

	return ra.Rollapp.InitialSequencer, nil
}

func IsInitialSequencer(addr, raID string, hd consts.HubData) (bool, error) {
	initSeqAddr, err := GetInitialSequencerAddress(raID, hd)
	if err != nil {
		return false, err
	}

	fmt.Printf("%s\n%s\n", addr, initSeqAddr)

	if strings.TrimSpace(addr) == strings.TrimSpace(initSeqAddr) {
		return true, nil
	}

	return false, nil
}

// TODO: most of rollapp utility functions should be tied to an entity
func IsRegistered(raID string, hd consts.HubData) (bool, error) {
	cmd := GetShowRollappCmd(raID, hd)
	_, err := bashutils.ExecCommandWithStdout(cmd)
	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			return false, errors.New("rollapp not found ")
		}
		return false, err
	}

	return true, nil
}

func GetShowRollappCmd(raID string, hd consts.HubData) *exec.Cmd {
	cmd := exec.Command(
		consts.Executables.Dymension,
		"q",
		"rollapp",
		"show",
		raID,
		"-o", "json",
		"--node", hd.RpcUrl,
		"--chain-id", hd.ID,
	)

	return cmd
}

func GetRollappCmd(raID string, hd consts.HubData) *exec.Cmd {
	cmd := exec.Command(
		consts.Executables.Dymension,
		"q", "rollapp", "show",
		raID, "-o", "json", "--node", hd.RpcUrl, "--chain-id", hd.ID,
	)

	return cmd
}

type GetProposerResponse struct {
	ProposerAddr string `json:"proposerAddr"`
}

func GetCurrentProposerCmd(raID string, hd consts.HubData) *exec.Cmd {
	cmd := exec.Command(
		consts.Executables.Dymension,
		"q", "sequencer", "proposer",
		raID, "-o", "json", "--node", hd.RpcUrl, "--chain-id", hd.ID,
	)

	return cmd
}

func GetCurrentProposer(raID string, hd consts.HubData) (string, error) {
	cmd := GetCurrentProposerCmd(raID, hd)

	out, err := bashutils.ExecCommandWithStdout(cmd)
	if err != nil {
		return "", err
	}
	var resp GetProposerResponse

	err = json.Unmarshal(out.Bytes(), &resp)
	if err != nil {
		return "", err
	}

	return resp.ProposerAddr, nil
}

func RollappConfigDir(root string) string {
	return filepath.Join(root, consts.ConfigDirName.Rollapp, "config")
}

func GetRollappSequencerAddress(home string) (string, error) {
	seqKeyConfig := keys.KeyConfig{
		Dir:         consts.ConfigDirName.Rollapp,
		ID:          consts.KeysIds.RollappSequencer,
		ChainBinary: consts.Executables.RollappEVM,
		Type:        consts.EVM_ROLLAPP,
	}

	addr, err := seqKeyConfig.Address(home)
	if err != nil {
		return "", err
	}

	return addr, nil
}

func GetMetadataFromChain(
	raID string,
	hd consts.HubData,
) (*ShowRollappResponse, error) {
	var raResponse ShowRollappResponse
	getRollappCmd := GetRollappCmd(raID, hd)

	out, err := bash.ExecCommandWithStdout(getRollappCmd)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(out.Bytes(), &raResponse)
	if err != nil {
		return nil, err
	}

	return &raResponse, nil
}

func PopulateRollerConfigWithRaMetadataFromChain(
	home, raID string,
	hd consts.HubData,
) (*roller.RollappConfig, error) {
	var cfg roller.RollappConfig
	raResponse, err := GetMetadataFromChain(raID, hd)
	if err != nil {
		return nil, err
	}

	vmt, _ := consts.ToVMType(strings.ToLower(raResponse.Rollapp.VmType))
	var kb consts.SupportedKeyringBackend

	rollerConfigExists, err := filesystem.DoesFileExist(roller.GetConfigPath(home))
	if err != nil {
		return nil, err
	}

	if rollerConfigExists {
		pterm.Info.Println(
			"existing roller configuration found, retrieving keyring backend from it",
		)
		rollerData, err := roller.LoadConfig(home)
		if err != nil {
			pterm.Error.Printf("failed to load roller config: %v\n", err)
			return nil, err
		}
		if rollerData.KeyringBackend == "" {
			pterm.Info.Println(
				"keyring backend not set in roller config, retrieving it from environment",
			)
			kb = keys.KeyringBackendFromEnv(hd.Environment)
		} else {
			kb = rollerData.KeyringBackend
		}
	} else {
		pterm.Info.Println("no existing roller configuration found, retrieving keyring backend from environment")
		kb = keys.KeyringBackendFromEnv(hd.Environment)
	}

	var DA consts.DaData

	switch hd.ID {
	case consts.MockHubID:
	case consts.MainnetHubID:
		DA = consts.DaNetworks[string(consts.CelestiaMainnet)]
	default:
		DA = consts.DaNetworks[string(consts.CelestiaTestnet)]
	}

	var baseDenom string
	if raResponse.Rollapp.GenesisInfo.NativeDenom != nil &&
		raResponse.Rollapp.GenesisInfo.NativeDenom.Base != "" {
		baseDenom = raResponse.Rollapp.GenesisInfo.NativeDenom.Base
	} else {
		pterm.Info.Println("no native denom found, assuming token-less rollapp and using the staking denom")
		genesisTmpDir, err := os.MkdirTemp(os.TempDir(), "genesis-file")
		if err != nil {
			return nil, err
		}
		// nolint: errcheck
		defer os.RemoveAll(genesisTmpDir)

		err = downloadGenesis(genesisTmpDir, raResponse.Rollapp.Metadata.GenesisUrl)
		if err != nil {
			pterm.Error.Println("failed to download genesis file: ", err)
			return nil, err
		}
		as, err := GetAppStateFromGenesisFile(genesisTmpDir)
		if err != nil {
			return nil, err
		}
		baseDenom = as.Staking.Params.BondDenom
	}

	cfg = roller.RollappConfig{
		Home:                 home,
		KeyringBackend:       kb,
		GenesisHash:          raResponse.Rollapp.GenesisInfo.GenesisChecksum,
		GenesisUrl:           raResponse.Rollapp.Metadata.GenesisUrl,
		RollappID:            raResponse.Rollapp.RollappId,
		RollappBinary:        consts.Executables.RollappEVM,
		RollappVMType:        vmt,
		Denom:                baseDenom,
		Decimals:             18,
		HubData:              hd,
		DA:                   DA,
		RollerVersion:        version.BuildCommit,
		Environment:          hd.ID,
		RollappBinaryVersion: version.BuildVersion,
		Bech32Prefix:         raResponse.Rollapp.GenesisInfo.Bech32Prefix,
		BaseDenom:            baseDenom,
		MinGasPrices:         "0",
	}

	return &cfg, nil
}

func Show(raID string, hd consts.HubData) (*ShowRollappResponse, error) {
	getRaCmd := GetRollappCmd(raID, hd)
	var raResponse ShowRollappResponse

	out, err := bashutils.ExecCommandWithStdout(getRaCmd)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(out.Bytes(), &raResponse)
	if err != nil {
		return nil, err
	}

	return &raResponse, nil
}

type RaParams struct {
	Params MinSequencerBond `json:"params"`
}

type MinSequencerBond struct {
	MinSequencerBondGlobal cosmossdktypes.Coin `json:"min_sequencer_bond_global"`
}

func GetRollappParams(hd consts.HubData) (*RaParams, error) {
	var resp RaParams
	cmd := exec.Command(
		consts.Executables.Dymension,
		"q",
		"rollapp",
		"params",
		"--node",
		hd.RpcUrl,
		"--chain-id",
		hd.ID,
		"-o",
		"json",
	)

	out, err := bash.ExecCommandWithStdout(cmd)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(out.Bytes(), &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// this is a duplicate of the one in genesisutils, a quick fix to make things work
func downloadGenesis(home, genesisUrl string) error {
	genesisPath := getGenesisFilePath(home)
	if genesisUrl == "" {
		return fmt.Errorf("RollApp's genesis url field is empty, contact the rollapp owner")
	}

	err := filesystem.DownloadFile(genesisUrl, genesisPath)
	if err != nil {
		return err
	}

	return nil
}

func getGenesisFilePath(root string) string {
	return filepath.Join(
		RollappConfigDir(root),
		"genesis.json",
	)
}

func GetAppStateFromGenesisFile(home string) (*AppState, error) {
	genesisFile, err := os.Open(getGenesisFilePath(home))
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
