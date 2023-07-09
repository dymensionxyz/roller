package celestia

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os/exec"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
)

// TODO: test how much is enough to run the LC for one day and set the minimum balance accordingly.
const (
	gatewayAddr             = "0.0.0.0"
	gatewayPort             = "26659"
	CelestiaRestApiEndpoint = "https://api-arabica-9.consensus.celestia-arabica.com"
	DefaultCelestiaRPC      = "consensus-full-arabica-9.celestia-arabica.com"
	DefaultCelestiaNetwork  = "arabica"
)

var (
	lcMinBalance = big.NewInt(1)
	LCEndpoint   = fmt.Sprintf("http://%s:%s", gatewayAddr, gatewayPort)
)

type Celestia struct {
	Root        string
	rpcEndpoint string
}

func (c2 *Celestia) GetStatus(c config.RollappConfig) string {
	logger := utils.GetRollerLogger(c.Home)
	out, err := utils.RestQueryJson(fmt.Sprintf("%s/balance", consts.DefaultDALCRPC))
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
	return LCEndpoint
}

// GetDAAccountAddress implements datalayer.DataLayer.
func (c *Celestia) GetDAAccountAddress() (string, error) {
	daKeysDir := filepath.Join(c.Root, consts.ConfigDirName.DALightNode, consts.KeysDirName)
	cmd := exec.Command(
		consts.Executables.CelKey, "show", consts.KeysIds.DALightNode, "--node.type", "light", "--keyring-dir",
		daKeysDir, "--keyring-backend", "test", "--output", "json",
	)
	output, err := utils.ExecBashCommand(cmd)
	if err != nil {
		return "", err
	}
	address, err := utils.ParseAddressFromOutput(output)
	return address, err
}

// TODO: wrap in some DA interfafce to be used for Avail as well
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
	return []utils.AccountData{*celAddress}, err
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
			KeyName:         consts.KeysIds.DALightNode,
			Denom:           consts.Denoms.Celestia,
			Network:         DefaultCelestiaNetwork,
		})
	}
	return insufficientBalances, nil
}

func (c *Celestia) GetStartDACmd() *exec.Cmd {
	return exec.Command(
		consts.Executables.Celestia, "light", "start",
		"--core.ip", c.rpcEndpoint,
		"--node.store", filepath.Join(c.Root, consts.ConfigDirName.DALightNode),
		"--gateway",
		"--gateway.deprecated-endpoints",
		"--gateway.addr", gatewayAddr,
		"--gateway.port", gatewayPort,
		"--p2p.network", DefaultCelestiaNetwork,
	)
}

func (c *Celestia) SetRPCEndpoint(rpc string) {
	c.rpcEndpoint = rpc
}

func (c *Celestia) GetNetworkName() string {
	return DefaultCelestiaNetwork
}
