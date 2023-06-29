package utils

import (
	"fmt"
	"io/ioutil"
	"math/big"
	"path/filepath"
	"regexp"

	"github.com/pelletier/go-toml"
)

func WriteConfigToTOML(InitConfig RollappConfig) error {
	tomlBytes, err := toml.Marshal(InitConfig)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(InitConfig.Home, RollerConfigFileName), tomlBytes, 0644)
	if err != nil {
		return err
	}

	return nil
}

func LoadConfigFromTOML(root string) (RollappConfig, error) {
	var config RollappConfig
	tomlBytes, err := ioutil.ReadFile(filepath.Join(root, RollerConfigFileName))
	if err != nil {
		return config, err
	}
	err = toml.Unmarshal(tomlBytes, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

type RollappConfig struct {
	Home          string
	RollappID     string
	RollappBinary string
	Denom         string
	TokenSupply   string
	HubData       HubData
}

func (c RollappConfig) Validate() error {
	err := VerifyHubID(c.HubData)
	if err != nil {
		return err
	}
	err = VerifyTokenSupply(c.TokenSupply)
	if err != nil {
		return err
	}
	err = ValidateRollAppID(c.RollappID)
	if err != nil {
		return fmt.Errorf("invalid RollApp ID '%s'", c.RollappID)
	}
	return nil
}

const RollerConfigFileName = "config.toml"

type HubData = struct {
	API_URL string
	ID      string
	RPC_URL string
}

func ValidateRollAppID(id string) error {
	pattern := `^[a-z]+_[0-9]{1,5}-[0-9]{1,5}$`
	r, _ := regexp.Compile(pattern)
	if !r.MatchString(id) {
		return fmt.Errorf("invalid RollApp ID '%s'", id)
	}
	return nil
}

func VerifyHubID(data HubData) error {
	if data.RPC_URL == "" {
		return fmt.Errorf("invalid hub ID: %s. RPC URL cannot be empty", data.ID)
	}
	return nil
}

func VerifyTokenSupply(supply string) error {
	tokenSupply := new(big.Int)
	_, ok := tokenSupply.SetString(supply, 10)
	if !ok {
		return fmt.Errorf("invalid token supply: %s. Must be a valid integer", supply)
	}

	ten := big.NewInt(10)
	remainder := new(big.Int)
	remainder.Mod(tokenSupply, ten)

	if remainder.Cmp(big.NewInt(0)) != 0 {
		return fmt.Errorf("invalid token supply: %s. Must be divisible by 10", supply)
	}

	return nil
}
