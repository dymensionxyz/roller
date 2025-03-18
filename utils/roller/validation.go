package roller

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/dymensionxyz/roller/cmd/consts"
)

func IsValidDAType(t string) bool {
	switch consts.DAType(t) {
	case consts.Local, consts.Celestia, consts.Avail, consts.WeaveVM, consts.Near:
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

	if data.RpcUrl == "" {
		return fmt.Errorf("invalid RPC endpoint: %s. RPC URL cannot be empty", data.ID)
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
