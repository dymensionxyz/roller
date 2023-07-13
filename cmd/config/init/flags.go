package initconfig

import (
	"fmt"
	"strings"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"github.com/spf13/cobra"
)

const (
	defaultTokenSupply = "1000000000"
)

func addFlags(cmd *cobra.Command) error {
	cmd.Flags().StringP(FlagNames.HubID, "", StagingHubName, fmt.Sprintf("The ID of the Dymension hub. %s", getAvailableHubsMessage()))
	cmd.Flags().StringP(FlagNames.RollappBinary, "", consts.Executables.RollappEVM, "The rollapp binary. Should be passed only if you built a custom rollapp")
	cmd.Flags().StringP(FlagNames.TokenSupply, "", defaultTokenSupply, "The total token supply of the RollApp")
	cmd.Flags().BoolP(FlagNames.Interactive, "i", false, "Run roller in interactive mode")
	cmd.Flags().UintP(FlagNames.Decimals, "", 18,
		"The precision level of the RollApp's token defined by the number of decimal places. "+
			"It should be an integer ranging between 1 and 18. This is akin to how 1 Ether equates to 10^18 Wei in Ethereum. "+
			"Note: EVM RollApps must set this value to 18.")

	cmd.Flags().StringP(FlagNames.DAType, "", "Celestia", "The DA layer for the RollApp. Can be one of 'Celestia, Avail, Mock'")

	// TODO: Expose when supporting custom sdk rollapps.
	err := cmd.Flags().MarkHidden(FlagNames.Decimals)
	if err != nil {
		return err
	}
	return nil
}

func GetInitConfig(initCmd *cobra.Command, args []string) (config.RollappConfig, error) {
	cfg := config.RollappConfig{}
	cfg.Home = initCmd.Flag(utils.FlagNames.Home).Value.String()
	cfg.RollappBinary = initCmd.Flag(FlagNames.RollappBinary).Value.String()
	// Error is ignored because the flag is validated in the cobra preRun hook
	decimals, _ := initCmd.Flags().GetUint(FlagNames.Decimals)
	cfg.Decimals = decimals
	interactive, _ := initCmd.Flags().GetBool(FlagNames.Interactive)
	if interactive {
		if err := RunInteractiveMode(&cfg); err != nil {
			return cfg, err
		}
		return cfg, nil
	}

	rollappId := args[0]
	denom := args[1]

	hubID := initCmd.Flag(FlagNames.HubID).Value.String()
	tokenSupply := initCmd.Flag(FlagNames.TokenSupply).Value.String()
	cfg.RollappID = rollappId
	cfg.Denom = "u" + denom
	cfg.HubData = Hubs[hubID]
	cfg.TokenSupply = tokenSupply
	cfg.DA = config.DAType(strings.ToLower(initCmd.Flag(FlagNames.DAType).Value.String()))

	return cfg, nil
}

func getAvailableHubsMessage() string {
	return fmt.Sprintf("Acceptable values are '%s' or '%s'", StagingHubName, LocalHubName)
}
