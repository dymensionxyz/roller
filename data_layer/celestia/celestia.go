package celestia

import (
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
	gatewayAddr = "0.0.0.0"
	gatewayPort = "26659"
)

var (
	lcMinBalance = big.NewInt(1)
	LCEndpoint   = fmt.Sprintf("http://%s:%s", gatewayAddr, gatewayPort)
)

type Celestia struct {
	Root string
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
	initLightNodeCmd := exec.Command(consts.Executables.Celestia, "light", "init", "--p2p.network", consts.DefaultCelestiaNetwork, "--node.store", filepath.Join(c.Root, consts.ConfigDirName.DALightNode))
	err := initLightNodeCmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (c *Celestia) GetDAAccData(config.RollappConfig) (*utils.AccountData, error) {
	celAddress, err := c.GetDAAccountAddress()
	if err != nil {
		return nil, err
	}
	var restQueryUrl = fmt.Sprintf(
		"%s/cosmos/bank/v1beta1/balances/%s",
		consts.CelestiaRestApiEndpoint, celAddress,
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

func (c *Celestia) CheckDABalance() ([]utils.NotFundedAddressData, error) {
	accData, err := c.GetDAAccData(config.RollappConfig{})
	if err != nil {
		return nil, err
	}
	var insufficientBalances []utils.NotFundedAddressData
	if accData.Balance.Cmp(lcMinBalance) < 0 {
		insufficientBalances = append(insufficientBalances, utils.NotFundedAddressData{
			Address:         accData.Address,
			CurrentBalance:  accData.Balance,
			RequiredBalance: lcMinBalance,
			KeyName:         consts.KeysIds.DALightNode,
			Denom:           consts.Denoms.Celestia,
		})
	}
	return insufficientBalances, nil
}

func (c *Celestia) GetStartDACmd(rpcEndpoint string) *exec.Cmd {
	return exec.Command(
		consts.Executables.Celestia, "light", "start",
		"--core.ip", rpcEndpoint,
		"--node.store", filepath.Join(c.Root, consts.ConfigDirName.DALightNode),
		"--gateway",
		"--gateway.addr", gatewayAddr,
		"--gateway.port", gatewayPort,
		"--p2p.network", consts.DefaultCelestiaNetwork,
	)
}
