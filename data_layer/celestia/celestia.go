package celestia

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	globalutils "github.com/dymensionxyz/roller/utils"
)

// TODO: test how much is enough to run the LC for one day and set the minimum balance accordingly.
const (
	CelestiaRestApiEndpoint = "https://api-arabica-9.consensus.celestia-arabica.com"
	DefaultCelestiaRPC      = "consensus-full-arabica-9.celestia-arabica.com"
	DefaultCelestiaNetwork  = "arabica"
)

var (
	lcMinBalance = big.NewInt(1)
)

type Celestia struct {
	Root            string
	rpcEndpoint     string
	metricsEndpoint string
	RPCPort         string
}

var instance *Celestia
var once sync.Once

func GetInstance(home string) *Celestia {
	once.Do(func() {
		cel := &Celestia{
			Root: home,
		}
		err := cel.InitializeLightNodeConfig()
		utils.PrettifyErrorIfExists(err)
		rpcPort, err := cel.ReadLightNodePort()
		if err != nil {
			panic(err)
		}
		cel.RPCPort = rpcPort
		instance = cel
	})
	return instance
}

func (c2 *Celestia) GetPrivateKey() (string, error) {
	exportKeyCmd := c2.GetExportKeyCmd()
	out, err := utils.ExecBashCommandWithStdErr(exportKeyCmd)
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func (c2 *Celestia) SetMetricsEndpoint(endpoint string) {
	c2.metricsEndpoint = endpoint
}

func (c *Celestia) GetStatus(rlpCfg config.RollappConfig) string {
	logger := utils.GetRollerLogger(rlpCfg.Home)
	lcEndpoint := c.GetLightNodeEndpoint()
	out, err := utils.RestQueryJson(fmt.Sprintf("%s/balance", lcEndpoint))
	const stoppedMsg = "Stopped, Restarting..."
	if err != nil {
		logger.Println("Error querying balance", err)
		return stoppedMsg
	} else {
		var balanceResp utils.BalanceResp
		err := json.Unmarshal(out.Bytes(), &balanceResp)
		if err != nil {
			logger.Println("Error unmarshalling balance response", err)
			return stoppedMsg
		}
		balance, err := utils.ParseBalance(balanceResp)
		if err != nil {
			logger.Println("Error parsing balance", err)
			return stoppedMsg
		}
		if balance.Cmp(lcMinBalance) < 0 {
			return "Stopped, out of funds"
		}
		return "Active"
	}
}

func (c *Celestia) GetLightNodeEndpoint() string {
	return fmt.Sprintf("http://localhost:%s", c.RPCPort)
}

func (c *Celestia) ReadLightNodePort() (string, error) {
	port, err := globalutils.GetKeyFromTomlFile(filepath.Join(c.Root, consts.ConfigDirName.DALightNode, "config.toml"),
		"Gateway.Port")
	if err != nil {
		return "", err
	}
	return port, nil
}

// GetDAAccountAddress implements datalayer.DataLayer.
// FIXME: should be loaded once and cached
func (c *Celestia) GetDAAccountAddress() (string, error) {
	daKeysDir := filepath.Join(c.Root, consts.ConfigDirName.DALightNode, consts.KeysDirName)
	cmd := exec.Command(
		consts.Executables.CelKey, "show", c.GetKeyName(), "--node.type", "light", "--keyring-dir",
		daKeysDir, "--keyring-backend", "test", "--output", "json",
	)
	output, err := utils.ExecBashCommandWithStdout(cmd)
	if err != nil {
		return "", err
	}
	address, err := utils.ParseAddressFromOutput(output)
	return address, err
}

func (c *Celestia) InitializeLightNodeConfig() error {
	initLightNodeCmd := exec.Command(consts.Executables.Celestia, "light", "init", "--p2p.network",
		DefaultCelestiaNetwork, "--node.store", filepath.Join(c.Root, consts.ConfigDirName.DALightNode))
	err := initLightNodeCmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (c *Celestia) getDAAccData(config.RollappConfig) (*utils.AccountData, error) {
	celAddress, err := c.GetDAAccountAddress()
	if err != nil {
		return nil, err
	}
	var restQueryUrl = fmt.Sprintf(
		"%s/cosmos/bank/v1beta1/balances/%s",
		CelestiaRestApiEndpoint, celAddress,
	)
	balancesJson, err := utils.RestQueryJson(restQueryUrl)
	if err != nil {
		return nil, err
	}
	balance, err := utils.ParseBalanceFromResponse(*balancesJson, consts.Denoms.Celestia)
	if err != nil {
		return nil, err
	}
	return &utils.AccountData{
		Address: celAddress,
		Balance: balance,
	}, nil
}

func (c *Celestia) GetDAAccData(cfg config.RollappConfig) ([]utils.AccountData, error) {
	celAddress, err := c.getDAAccData(cfg)
	if err != nil {
		return nil, err
	}
	if celAddress == nil {
		return nil, fmt.Errorf("failed to get DA account data")
	}
	return []utils.AccountData{*celAddress}, err
}

func (c *Celestia) GetKeyName() string {
	return "my_celes_key"
}

func (c *Celestia) GetExportKeyCmd() *exec.Cmd {
	return utils.GetExportKeyCmdBinary(c.GetKeyName(), filepath.Join(c.Root, consts.ConfigDirName.DALightNode, "keys"),
		consts.Executables.CelKey)
}

func (c *Celestia) CheckDABalance() ([]utils.NotFundedAddressData, error) {
	accData, err := c.getDAAccData(config.RollappConfig{})
	if err != nil {
		return nil, err
	}
	var insufficientBalances []utils.NotFundedAddressData
	if accData.Balance.Amount.Cmp(lcMinBalance) < 0 {
		insufficientBalances = append(insufficientBalances, utils.NotFundedAddressData{
			Address:         accData.Address,
			CurrentBalance:  accData.Balance.Amount,
			RequiredBalance: lcMinBalance,
			KeyName:         c.GetKeyName(),
			Denom:           consts.Denoms.Celestia,
			Network:         DefaultCelestiaNetwork,
		})
	}
	return insufficientBalances, nil
}

func (c *Celestia) GetStartDACmd() *exec.Cmd {
	args := []string{
		"light", "start",
		"--core.ip", c.rpcEndpoint,
		"--node.store", filepath.Join(c.Root, consts.ConfigDirName.DALightNode),
		"--gateway",
		"--gateway.deprecated-endpoints",
		"--p2p.network", DefaultCelestiaNetwork,
	}
	if c.metricsEndpoint != "" {
		args = append(args, "--metrics", "--metrics.endpoint", c.metricsEndpoint)
	}
	return exec.Command(
		consts.Executables.Celestia, args...,
	)
}

func (c *Celestia) SetRPCEndpoint(rpc string) {
	c.rpcEndpoint = rpc
}

func (c *Celestia) GetNetworkName() string {
	return DefaultCelestiaNetwork
}

func (c *Celestia) GetSequencerDAConfig() string {
	lcEndpoint := c.GetLightNodeEndpoint()
	return fmt.Sprintf(`{"base_url": "%s", "timeout": 60000000000, "gas_prices":0.1, "gas_adjustment": 1.3, "namespace_id":"000000000000ffff"}`,
		lcEndpoint)
}

// UnsafeResetInstance resets the singleton instance. Use only in tests.
func UnsafeResetInstance() {
	instance = nil
	once = sync.Once{}
}
