package initconfig

import (
	"fmt"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	global_utils "github.com/dymensionxyz/roller/utils"
	"github.com/pelletier/go-toml"
	"os/exec"
	"path/filepath"
)

const validatorKeyID = "local-user"

func initLocalHub(rlpCfg config.RollappConfig) error {
	initBashCmd := getInitDymdCmd(rlpCfg)
	_, err := utils.ExecBashCommandWithStdout(initBashCmd)
	if err != nil {
		return err
	}
	localHubPath := filepath.Join(rlpCfg.Home, consts.ConfigDirName.LocalHub)
	if err = UpdateJSONParams(filepath.Join(localHubPath, "config", "genesis.json"), getHubGenesisParams()); err != nil {
		return err
	}
	if err := UpdateTendermintConfig(rlpCfg); err != nil {
		return err
	}
	if err := UpdateAppConfig(rlpCfg); err != nil {
		return err
	}
	if err := UpdateClientConfig(rlpCfg); err != nil {
		return err
	}
	if err != nil {
		return err
	}
	addr, err := createAddressBinary(utils.KeyConfig{
		Dir:         consts.ConfigDirName.LocalHub,
		ID:          validatorKeyID,
		ChainBinary: consts.Executables.Dymension,
		Type:        config.SDK_ROLLAPP,
	}, rlpCfg.Home)
	if err != nil {
		return err
	}
	addGenAccountCmd := exec.Command(consts.Executables.Dymension, "add-genesis-account", addr,
		"1000000000000000000000000"+consts.Denoms.Hub, "--home", localHubPath)
	_, err = utils.ExecBashCommandWithStdout(addGenAccountCmd)
	if err != nil {
		return err
	}
	genTxCmd := exec.Command(consts.Executables.Dymension, "gentx", validatorKeyID, "670000000000000000000000"+consts.Denoms.Hub,
		"--home", localHubPath, "--chain-id", rlpCfg.HubData.ID, "--keyring-backend", "test")
	_, err = utils.ExecBashCommandWithStdout(genTxCmd)
	if err != nil {
		return err
	}
	collectGentxsCmd := exec.Command(consts.Executables.Dymension, "collect-gentxs", "--home",
		localHubPath)
	_, err = utils.ExecBashCommandWithStdout(collectGentxsCmd)
	if err != nil {
		return err
	}
	return nil
}

func getInitDymdCmd(rlpCfg config.RollappConfig) *exec.Cmd {
	return exec.Command(consts.Executables.Dymension, "init", "local", "--chain-id", rlpCfg.HubData.ID, "--home",
		filepath.Join(rlpCfg.Home, consts.ConfigDirName.LocalHub))
}

func getHubGenesisParams() []PathValue {
	return []PathValue{
		{"app_state.gov.deposit_params.min_deposit.0.denom", consts.Denoms.Hub},
		{"app_state.gov.deposit_params.min_deposit.0.amount", "10000000000"},
		{"app_state.gov.voting_params.voting_period", "30s"},
		{"app_state.rollapp.params.dispute_period_in_blocks", "2"},
		{"app_state.staking.params.max_validators", "110"},
		{"consensus_params.block.max_gas", "40000000"},
		{"app_state.feemarket.params.no_base_fee", true},
		{"app_state.evm.params.evm_denom", consts.Denoms.Hub},
		{"app_state.evm.params.enable_create", false},
		{"app_state.crisis.constant_fee.denom", consts.Denoms.Hub},
		{"app_state.staking.params.bond_denom", consts.Denoms.Hub},
		{"app_state.mint.params.mint_denom", consts.Denoms.Hub},
	}
}

func UpdateTendermintConfig(rlpCfg config.RollappConfig) error {
	tendermintConfigFilePath := filepath.Join(rlpCfg.Home, consts.ConfigDirName.LocalHub, "config", "config.toml")
	tmCfg, err := toml.LoadFile(tendermintConfigFilePath)
	if err != nil {
		return fmt.Errorf("failed to load %s: %v", tendermintConfigFilePath, err)
	}
	tmCfg.Set("rpc.laddr", "tcp://0.0.0.0:36657")
	tmCfg.Set("p2p.laddr", "tcp://0.0.0.0:36656")
	tmCfg.Set("prometheus", true)
	tmCfg.Set("cors_allowed_origins", []string{"*"})
	return global_utils.WriteTomlTreeToFile(tmCfg, tendermintConfigFilePath)
}

func UpdateAppConfig(rlpCfg config.RollappConfig) error {
	appConfigFilePath := filepath.Join(rlpCfg.Home, consts.ConfigDirName.LocalHub, "config", "app.toml")
	appCfg, err := toml.LoadFile(appConfigFilePath)
	if err != nil {
		return fmt.Errorf("failed to load %s: %v", appConfigFilePath, err)
	}
	appCfg.Set("grpc.address", "0.0.0.0:8090")
	appCfg.Set("grpc-web.address", "0.0.0.0:8091")
	appCfg.Set("json-rpc.address", "0.0.0.0:9545")
	appCfg.Set("json-rpc.ws-address", "0.0.0.0:9546")
	appCfg.Set("api.enable", true)
	appCfg.Set("api.address", "tcp://0.0.0.0:1318")
	appCfg.Set("minimum-gas-prices", "0"+consts.Denoms.Hub)
	appCfg.Set("telemetry.enabled", true)
	appCfg.Set("prometheus-retention-time", "31104000")
	appCfg.Set("enabled-unsafe-cors", true)
	return global_utils.WriteTomlTreeToFile(appCfg, appConfigFilePath)
}

func UpdateClientConfig(rlpCfg config.RollappConfig) error {
	clientConfigFilePath := filepath.Join(rlpCfg.Home, consts.ConfigDirName.LocalHub, "config", "client.toml")
	clientCfg, err := toml.LoadFile(clientConfigFilePath)
	if err != nil {
		return fmt.Errorf("failed to load %s: %v", clientConfigFilePath, err)
	}
	clientCfg.Set("chain-id", rlpCfg.HubData.ID)
	clientCfg.Set("keyring-backend", "test")
	clientCfg.Set("node", "tcp://0.0.0.0:36657")
	return global_utils.WriteTomlTreeToFile(clientCfg, clientConfigFilePath)
}
