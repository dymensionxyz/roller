package celestia

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os/exec"
	"path/filepath"
	"strings"

	cosmossdkmath "cosmossdk.io/math"
	cosmossdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/roller"
)

var lcMinBalance = big.NewInt(1)

type Celestia struct {
	Root            string
	rpcEndpoint     string
	metricsEndpoint string
	RPCPort         string
	NamespaceID     string
	KeyringBackend  consts.SupportedKeyringBackend
}

func NewCelestia(home string, kb consts.SupportedKeyringBackend) *Celestia {
	return &Celestia{
		Root:           home,
		KeyringBackend: kb,
	}
}

func (c *Celestia) GetPrivateKey() (string, error) {
	exportKeyCmd := c.GetExportKeyCmd()
	out, err := bash.ExecCommandWithStdErr(exportKeyCmd)
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func (c *Celestia) SetMetricsEndpoint(endpoint string) {
	c.metricsEndpoint = endpoint
}

type BalanceResponse struct {
	Result cosmossdktypes.Coin `json:"result"`
}

func (c *Celestia) GetStatus(rlpCfg roller.RollappConfig) string {
	args := []string{
		"state",
		"balance",
		"--node.store",
		filepath.Join(c.Root, consts.ConfigDirName.DALightNode),
	}
	output, err := exec.Command(consts.Executables.Celestia, args...).Output()
	if err != nil {
		return "Stopped, Restarting..."
	}

	var resp BalanceResponse
	err = json.Unmarshal(output, &resp)
	if err != nil {
		return "Stopped, Restarting..."
	}

	if resp.Result.Amount != cosmossdkmath.NewInt(0) {
		return "active"
	}
	// if strings.TrimSpace(resp.Result.Amount) != 0 {
	// 	return "active"
	// }

	return "Stopped, Restarting..."
}

func (c *Celestia) GetRootDirectory() string {
	return c.Root
}

func (c *Celestia) getRPCPort() string {
	if c.RPCPort != "" {
		return c.RPCPort
	}
	port, err := tomlconfig.GetKeyFromFile(
		filepath.Join(c.Root, consts.ConfigDirName.DALightNode, "config.toml"),
		"RPC.Port",
	)
	if err != nil {
		panic(err)
	}
	c.RPCPort = port
	return port
}

func (c *Celestia) GetLightNodeEndpoint() string {
	return fmt.Sprintf("http://localhost:%s", c.getRPCPort())
}

// GetDAAccountAddress implements datalayer.DataLayer.
// FIXME: should be loaded once and cached
func (c *Celestia) GetDAAccountAddress() (*keys.KeyInfo, error) {
	daKeysDir := filepath.Join(c.Root, consts.ConfigDirName.DALightNode, consts.KeysDirName)

	args := []string{
		"show", c.GetKeyName(), "--node.type", "light", "--keyring-dir",
		daKeysDir, "--keyring.backend", string(c.KeyringBackend), "--output", "json",
	}
	cmd := exec.Command(
		consts.Executables.CelKey, args...,
	)

	fmt.Println("cmd:", cmd.String())

	if c.KeyringBackend == consts.SupportedKeyringBackends.OS {
		psw, err := config.ReadFromFile(
			filepath.Join(c.Root, string(consts.OsKeyringPwdFileNames.Da)),
		)
		if err != nil {
			return nil, err
		}
		cmd.Env = append(cmd.Env, fmt.Sprintf("DA_KEYRING_PASSPHRASE=%s", psw))
	}
	output, err := bash.ExecCommandWithStdout(cmd)
	if err != nil {
		return nil, err
	}

	address, err := keys.ParseAddressFromOutput(output)
	return address, err
}

func (c *Celestia) InitializeLightNodeConfig() (string, error) {
	raCfg, err := roller.LoadConfig(c.Root)
	if err != nil {
		return "", err
	}

	if c.KeyringBackend == consts.SupportedKeyringBackends.OS {
		pterm.Info.Println("creating keyring passphrase file")
		err := keys.CreateDaOsKeyringPswFile(c.Root)
		if err != nil {
			return "", err
		}
	}

	args := []string{
		"light", "init",
		"--p2p.network",
		string(raCfg.DA.ID),
		"--node.store", filepath.Join(c.Root, consts.ConfigDirName.DALightNode),
		"--keyring.backend", string(c.KeyringBackend),
	}
	initLightNodeCmd := exec.Command(
		consts.Executables.Celestia, args...,
	)
	fmt.Println("initLightNodeCmd:", initLightNodeCmd.String())
	// err := initLightNodeCmd.Run()

	var out *bytes.Buffer
	out, err = keys.RunCmdBasedOnKeyringBackend(
		c.Root,
		consts.Executables.Celestia,
		args,
		c.KeyringBackend,
	)
	if err != nil {
		return "", err
	}

	mnemonic := extractMnemonic(out.String())

	return mnemonic, nil
}

func extractMnemonic(output string) string {
	scanner := bufio.NewScanner(strings.NewReader(output))
	mnemonicLineFound := false
	var mnemonicLines []string

	for scanner.Scan() {
		line := scanner.Text()
		if mnemonicLineFound {
			// Collect all subsequent lines as part of the mnemonic
			mnemonicLines = append(mnemonicLines, line)
		}
		if strings.HasPrefix(line, "MNEMONIC") {
			mnemonicLineFound = true
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading output:", err)
		return ""
	}

	return strings.Join(mnemonicLines, " ")
}

func (c *Celestia) getDAAccData(home string) (*keys.AccountData, error) {
	celAddress, err := c.GetDAAccountAddress()
	if err != nil {
		return nil, err
	}

	// TODO: refactor to support multiple DA chains
	raCfg, err := roller.LoadConfig(home)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(
		consts.Executables.CelestiaApp,
		"q",
		"bank",
		"balances",
		celAddress.Address,
		"--node",
		raCfg.DA.RpcUrl,
		"--chain-id",
		string(raCfg.DA.ID),
		"-o", "json",
	)

	output, err := bash.ExecCommandWithStdout(cmd)
	if err != nil {
		return nil, err
	}
	b := bytes.NewBuffer(output.Bytes())

	balance, err := keys.ParseBalanceFromResponse(
		*b,
		consts.Denoms.Celestia,
	)
	if err != nil {
		return nil, err
	}
	return &keys.AccountData{
		Address: celAddress.Address,
		Balance: balance,
	}, nil
}

func (c *Celestia) GetDAAccData(cfg roller.RollappConfig) ([]keys.AccountData, error) {
	celAddress, err := c.getDAAccData(c.Root)
	if err != nil {
		return nil, err
	}
	if celAddress == nil {
		return nil, fmt.Errorf("failed to get DA account data")
	}
	return []keys.AccountData{*celAddress}, err
}

func (c *Celestia) GetKeyName() string {
	return consts.KeysIds.Celestia
}

func (c *Celestia) GetExportKeyCmd() *exec.Cmd {
	return keys.GetExportKeyCmdBinary(
		c.GetKeyName(),
		filepath.Join(c.Root, consts.ConfigDirName.DALightNode, "keys"),
		consts.Executables.CelKey,
		string(c.KeyringBackend),
	)
}

func (c *Celestia) CheckDABalance() ([]keys.NotFundedAddressData, error) {
	accData, err := c.getDAAccData(c.Root)
	if err != nil {
		return nil, err
	}

	raCfg, err := roller.LoadConfig(c.Root)
	if err != nil {
		return nil, err
	}

	var insufficientBalances []keys.NotFundedAddressData
	if accData.Balance.Amount.Cmp(lcMinBalance) < 0 {
		insufficientBalances = append(
			insufficientBalances, keys.NotFundedAddressData{
				Address:         accData.Address,
				CurrentBalance:  accData.Balance.Amount,
				RequiredBalance: lcMinBalance,
				KeyName:         c.GetKeyName(),
				Denom:           consts.Denoms.Celestia,
				Network:         string(raCfg.DA.ID),
			},
		)
	}

	return insufficientBalances, nil
}

func (c *Celestia) GetStartDACmd() *exec.Cmd {
	raCfg, err := roller.LoadConfig(c.Root)
	if err != nil {
		return nil
	}

	args := []string{
		"light", "start",
		"--core.ip", raCfg.DA.CurrentStateNode,
		"--node.store", filepath.Join(c.Root, consts.ConfigDirName.DALightNode),
		"--gateway",
		// "--gateway.deprecated-endpoints",
		"--p2p.network", string(raCfg.DA.ID),
	}
	if c.metricsEndpoint != "" {
		args = append(args, "--metrics", "--metrics.endpoint", c.metricsEndpoint)
	}
	startCmd := exec.Command(
		consts.Executables.Celestia, args...,
	)
	// startCmd.Env = append(os.Environ(), CUSTOM_ARABICA11_CONFIG)
	return startCmd
}

func (c *Celestia) SetRPCEndpoint(rpc string) {
	c.rpcEndpoint = rpc
}

func (c *Celestia) GetNetworkName() string {
	return consts.DefaultCelestiaNetwork
}

func (c *Celestia) GetNamespaceID() string {
	return c.NamespaceID
}

func (c *Celestia) getAuthToken(t string, raCfg roller.RollappConfig) (string, error) {
	getAuthTokenCmd := exec.Command(
		consts.Executables.Celestia,
		"light",
		"auth",
		t,
		"--p2p.network",
		string(raCfg.DA.ID),
		"--node.store",
		filepath.Join(c.Root, consts.ConfigDirName.DALightNode),
	)
	output, err := bash.ExecCommandWithStdout(getAuthTokenCmd)
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(output.String(), "\n"), nil
}

func (c *Celestia) GetSequencerDAConfig(nt string) string {
	if c.NamespaceID == "" {
		c.NamespaceID = generateRandNamespaceID()
	}
	lcEndpoint := c.GetLightNodeEndpoint()

	var authToken string
	var err error

	raCfg, err := roller.LoadConfig(c.Root)
	if err != nil {
		return ""
	}

	if nt == consts.NodeType.Sequencer {
		authToken, err = c.getAuthToken(consts.DaAuthTokenType.Admin, raCfg)
	} else if nt == consts.NodeType.FullNode {
		authToken, err = c.getAuthToken(consts.DaAuthTokenType.Read, raCfg)
	} else {
		// TODO: don't panic,return an err
		err := errors.New("invalid node type")
		panic(err)
	}

	if err != nil {
		panic(err)
	}

	return fmt.Sprintf(
		`{"base_url": "%s", "timeout": 60000000000, "gas_prices":0.02, "gas_adjustment": 1.3, "namespace_id":"%s","auth_token":"%s","backoff":{"initial_delay":6000000000,"max_delay":6000000000,"growth_factor":2},"retry_attempts":4,"retry_delay":3000000000}`,
		lcEndpoint,
		c.NamespaceID,
		authToken,
	)
}
