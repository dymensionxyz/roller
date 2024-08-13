package config

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/dymensionxyz/roller/cmd/consts"
)

var SupportedDas = []consts.DAType{consts.Celestia, consts.Avail, consts.Local}

type RollappConfig struct {
	Home          string        `toml:"home"`
	GenesisHash   string        `toml:"genesis_hash"`
	GenesisUrl    string        `toml:"genesis_url"`
	RollappID     string        `toml:"rollapp_id"`
	RollappBinary string        `toml:"rollapp_binary"`
	VMType        consts.VMType `toml:"execution"`
	Denom         string        `toml:"denom"`
	// TokenSupply   string
	Decimals      uint
	HubData       consts.HubData
	DA            consts.DAType
	RollerVersion string `toml:"roller_version"`

	// new roller.toml
	Environment string `toml:"environment"`
	// Execution        string `toml:"execution"`
	ExecutionVersion string `toml:"execution_version"`
	Bech32Prefix     string `toml:"bech32_prefix"`
	BaseDenom        string `toml:"base_denom"`
	MinGasPrices     string `toml:"minimum_gas_prices"`
}

func (c RollappConfig) Validate() error {
	err := VerifyHubData(c.HubData)
	if err != nil {
		return err
	}

	// the assumption is that the supply is coming from the genesis creator
	// err = VerifyTokenSupply(c.TokenSupply)
	// if err != nil {
	// 	return err
	// }

	err = IsValidDenom(c.BaseDenom)
	if err != nil {
		return err
	}
	if err := ValidateDecimals(c.Decimals); err != nil {
		return err
	}

	if !IsValidDAType(string(c.DA)) {
		return fmt.Errorf("invalid DA type: %s. supported types %s", c.DA, SupportedDas)
	}

	if !IsValidVMType(string(c.VMType)) {
		return fmt.Errorf("invalid VM type: %s", c.VMType)
	}

	return nil
}

func IsValidDAType(t string) bool {
	switch consts.DAType(t) {
	case consts.Local, consts.Celestia, consts.Avail:
		return true
	}
	return false
}

func IsValidVMType(t string) bool {
	switch consts.VMType(t) {
	case consts.SDK_ROLLAPP, consts.EVM_ROLLAPP:
		return true
	}
	return false
}

func VerifyHubData(data consts.HubData) error {
	if data.ID == "mock" {
		return nil
	}

	if data.ID == "" {
		return fmt.Errorf("invalid hub id: %s. ID cannot be empty", data.ID)
	}

	if data.RPC_URL == "" {
		return fmt.Errorf("invalid RPC endoint: %s. RPC URL cannot be empty", data.ID)
	}
	return nil
}

// func VerifyTokenSupply(supply string) error {
// 	tokenSupply := new(big.Int)
// 	_, ok := tokenSupply.SetString(supply, 10)
// 	if !ok {
// 		return fmt.Errorf("invalid token supply: %s. Must be a valid integer", supply)
// 	}
//
// 	ten := big.NewInt(10)
// 	remainder := new(big.Int)
// 	remainder.Mod(tokenSupply, ten)
//
// 	if remainder.Cmp(big.NewInt(0)) != 0 {
// 		return fmt.Errorf("invalid token supply: %s. Must be divisible by 10", supply)
// 	}
//
// 	if tokenSupply.Cmp(big.NewInt(10_000_000)) < 0 {
// 		return fmt.Errorf("token supply %s must be greater than 10,000,000", tokenSupply)
// 	}
//
// 	return nil
// }

func ValidateDecimals(decimals uint) error {
	if decimals > 18 {
		return fmt.Errorf("invalid decimals: %d. Must be less than or equal to 18", decimals)
	}
	return nil
}

func IsValidDenom(s string) error {
	if !strings.HasPrefix(s, "a") {
		return fmt.Errorf("invalid denom '%s'. denom expected to start with 'a'", s)
	}
	if !IsValidTokenSymbol(s[1:]) {
		return fmt.Errorf("invalid token symbol '%s'", s[1:])
	}
	return nil
}

func IsValidTokenSymbol(s string) bool {
	if len(s) < 3 || len(s) > 6 {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(r) ||
			!strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", r) {
			return false
		}
	}
	return true
}
