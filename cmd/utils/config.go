package utils

import (
	"fmt"
	"io/ioutil"
	"math/big"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

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
	Decimals      uint
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
		return err
	}
	err = IsValidDenom(c.Denom)
	if err != nil {
		return err
	}
	if err := ValidateDecimals(c.Decimals); err != nil {
		return err
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
	r := regexp.MustCompile(pattern)
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

func ValidateDecimals(decimals uint) error {
	if decimals > 18 {
		return fmt.Errorf("invalid decimals: %d. Must be less than or equal to 18", decimals)
	}
	return nil
}

func IsValidDenom(s string) error {
	if !strings.HasPrefix(s, "u") {
		return fmt.Errorf("invalid denom '%s'. denom expected to start with 'u'", s)
	}
	if !IsValidTokenSymbol(s[1:]) {
		return fmt.Errorf("invalid token symbol '%s'", s[1:])
	}
	return nil
}

func IsValidTokenSymbol(s string) bool {
	if len(s) != 3 {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(r) || !strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", r) {
			return false
		}
	}
	return true
}
