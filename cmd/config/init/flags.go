package initconfig

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"math/big"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/spf13/cobra"
)

const (
	defaultTokenSupply = "1000000000"
)

func addFlags(cmd *cobra.Command) error {
	cmd.Flags().StringP(FlagNames.HubID, "", StagingHubID, fmt.Sprintf("The ID of the Dymension hub. %s", getAvailableHubsMessage()))
	cmd.Flags().StringP(FlagNames.RollappBinary, "", "", "The rollapp binary. Should be passed only if you built a custom rollapp")
	cmd.Flags().StringP(FlagNames.TokenSupply, "", defaultTokenSupply, "The total token supply of the RollApp")
	cmd.Flags().UintP(FlagNames.Decimals, "", 18,
		"The precision level of the RollApp's token defined by the number of decimal places. "+
			"It should be an integer ranging between 1 and 18. This is akin to how 1 Ether equates to 10^18 Wei in Ethereum. "+
			"Note: EVM RollApps must set this value to 18.")

	// TODO: Expose when supporting custom sdk rollapps.
	err := cmd.Flags().MarkHidden(FlagNames.Decimals)
	if err != nil {
		return err
	}
	return nil
}

func getRollappBinaryPath(cmd *cobra.Command) string {
	rollappBinaryPath := cmd.Flag(FlagNames.RollappBinary).Value.String()
	if rollappBinaryPath == "" {
		return consts.Executables.RollappEVM
	}
	return rollappBinaryPath
}

func getTokenSupply(cmd *cobra.Command) string {
	return cmd.Flag(FlagNames.TokenSupply).Value.String()
}

func GetInitConfig(initCmd *cobra.Command, args []string) (utils.RollappConfig, error) {
	rollappId := args[0]
	denom := args[1]
	home := initCmd.Flag(utils.FlagNames.Home).Value.String()
	rollappBinaryPath := getRollappBinaryPath(initCmd)
	hubID := initCmd.Flag(FlagNames.HubID).Value.String()
	tokenSupply := getTokenSupply(initCmd)
	// Error is ignored because the flag is validated in the cobra preRun hook
	decimals, _ := initCmd.Flags().GetUint(FlagNames.Decimals)
	return utils.RollappConfig{
		Home:          home,
		RollappID:     rollappId,
		RollappBinary: rollappBinaryPath,
		Denom:         "u" + denom,
		HubData:       Hubs[hubID],
		TokenSupply:   tokenSupply,
		Decimals:      decimals,
	}, nil
}
func getValidRollappIdMessage() string {
	return "A valid RollApp ID should follow the format 'rollapp-name_EIP155-revision', where 'rollapp-name' is made up of" +
		" lowercase English letters, 'EIP155-revision' is a 1 to 5 digit number representing the EIP155 rollapp ID, and '" +
		"revision' is a 1 to 5 digit number representing the revision. For example: 'mars_9721-1'"
}

func getAvailableHubsMessage() string {
	return fmt.Sprintf("Acceptable values are '%s' or '%s'", StagingHubID, LocalHubID)
}

func validateRollAppID(id string) bool {
	pattern := `^[a-z]+_[0-9]{1,5}-[0-9]{1,5}$`
	r, _ := regexp.Compile(pattern)
	return r.MatchString(id)
}

func verifyHubID(cmd *cobra.Command) error {
	hubID, err := cmd.Flags().GetString(FlagNames.HubID)
	if err != nil {
		return err
	}
	if _, ok := Hubs[hubID]; !ok {
		return fmt.Errorf("invalid hub ID: %s. %s", hubID, getAvailableHubsMessage())
	}
	return nil
}

func verifyTokenSupply(cmd *cobra.Command) error {
	tokenSupplyStr, err := cmd.Flags().GetString(FlagNames.TokenSupply)
	if err != nil {
		return err
	}

	tokenSupply := new(big.Int)
	_, ok := tokenSupply.SetString(tokenSupplyStr, 10)
	if !ok {
		return fmt.Errorf("invalid token supply: %s. Must be a valid integer", tokenSupplyStr)
	}

	ten := big.NewInt(10)
	remainder := new(big.Int)
	remainder.Mod(tokenSupply, ten)

	if remainder.Cmp(big.NewInt(0)) != 0 {
		return fmt.Errorf("invalid token supply: %s. Must be divisible by 10", tokenSupplyStr)
	}

	return nil
}

func verifyDecimals(cmd *cobra.Command) error {
	decimals, err := cmd.Flags().GetUint(FlagNames.Decimals)
	if err != nil {
		return err
	}
	if decimals > 18 {
		return fmt.Errorf("invalid decimals: %d. Must be less than or equal to 18", decimals)
	}
	return nil
}

func isValidDenom(s string) bool {
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
