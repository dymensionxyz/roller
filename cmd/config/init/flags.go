package initconfig

import (
	"fmt"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/spf13/cobra"
)

const (
	defaultTokenSupply = "1000000000"
)

func addFlags(cmd *cobra.Command) error {
	cmd.Flags().StringP(FlagNames.HubID, "", StagingHubID, fmt.Sprintf("The ID of the Dymension hub. %s", getAvailableHubsMessage()))
	cmd.Flags().StringP(FlagNames.RollappBinary, "", consts.Executables.RollappEVM, "The rollapp binary. Should be passed only if you built a custom rollapp")
	cmd.Flags().StringP(FlagNames.TokenSupply, "", defaultTokenSupply, "The total token supply of the RollApp")
	cmd.Flags().BoolP(FlagNames.Interactive, "", false, "Run roller in interactive mode")
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

func getTokenSupply(cmd *cobra.Command) string {
	return cmd.Flag(FlagNames.TokenSupply).Value.String()
}

func GetInitConfig(initCmd *cobra.Command, args []string) (utils.RollappConfig, error) {
	cfg := utils.RollappConfig{}
	cfg.Home = initCmd.Flag(utils.FlagNames.Home).Value.String()
	cfg.RollappBinary = initCmd.Flag(FlagNames.RollappBinary).Value.String()
	// Error is ignored because the flag is validated in the cobra preRun hook
	decimals, _ := initCmd.Flags().GetUint(FlagNames.Decimals)
	cfg.Decimals = decimals
	interactive, _ := initCmd.Flags().GetBool(FlagNames.Interactive)
	if interactive {
		RunInteractiveMode(&cfg)
		return cfg, nil
	}

	rollappId := args[0]
	denom := args[1]

	hubID := initCmd.Flag(FlagNames.HubID).Value.String()
	tokenSupply := getTokenSupply(initCmd)
	cfg.RollappID = rollappId
	cfg.Denom = "u" + denom
	cfg.HubData = Hubs[hubID]
	cfg.TokenSupply = tokenSupply

	return cfg, nil
}
func getValidRollappIdMessage() string {
	return "A valid RollApp ID should follow the format 'name_EIP155-revision', where 'name' is made up of" +
		" lowercase English letters, 'EIP155-revision' is a 1 to 5 digit number representing the EIP155 rollapp ID, and '" +
		"revision' is a 1 to 5 digit number representing the revision. For example: 'mars_9721-1'"
}

func getAvailableHubsMessage() string {
	return fmt.Sprintf("Acceptable values are '%s' or '%s'", StagingHubID, LocalHubID)
}

func getValidDenomMessage() string {
	return "A valid denom should consist of exactly 3 English alphabet letters, for example 'btc', 'eth'"
}
