package initconfig

import (
	"fmt"
	"regexp"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/spf13/cobra"
	"math/big"
)

func addFlags(cmd *cobra.Command) {
	cmd.Flags().StringP(FlagNames.HubID, "", TestnetHubID, fmt.Sprintf("The ID of the Dymension hub. %s", getAvailableHubsMessage()))
	cmd.Flags().StringP(FlagNames.RollappBinary, "", "", "The rollapp binary. Should be passed only if you built a custom rollapp")
	cmd.Flags().StringP(FlagNames.TokenSupply, "", "1000000000", "The total token supply of the RollApp")

	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		err := verifyHubID(cmd)
		if err != nil {
			return err
		}
		err = verifyTokenSupply(cmd)
		if err != nil {
			return err
		}
		rollappID := args[0]
		if !validateRollAppID(rollappID) {
			return fmt.Errorf("invalid RollApp ID '%s'. %s", rollappID, getValidRollappIdMessage())
		}
		return nil
	}
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
	return utils.RollappConfig{
		Home:          home,
		RollappID:     rollappId,
		RollappBinary: rollappBinaryPath,
		Denom:         denom,
		HubData:       Hubs[hubID],
		TokenSupply:   tokenSupply,
	}, nil
}
func getValidRollappIdMessage() string {
	return "A valid RollApp ID should follow the format 'rollapp-name_EIP155_version', where 'rollapp-name' is made up of" +
		" lowercase English letters, 'EIP155_version' is a 1 to 5 digit number representing the EIP155 rollapp ID, and '" +
		"version' is a 1 to 5 digit number representing the version. For example: 'mars_9721_1'"
}

func getAvailableHubsMessage() string {
	return fmt.Sprintf("Acceptable values are '%s', '%s' or '%s'", TestnetHubID, StagingHubID, LocalHubID)
}

func validateRollAppID(id string) bool {
	pattern := `^[a-z]+_[0-9]{1,5}_[0-9]{1,5}$`
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
