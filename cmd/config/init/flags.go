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

func addFlags(cmd *cobra.Command) {
	cmd.Flags().StringP(FlagNames.HubID, "", StagingHubID, fmt.Sprintf("The ID of the Dymension hub. %s", getAvailableHubsMessage()))
	cmd.Flags().StringP(FlagNames.RollappBinary, "", "", "The rollapp binary. Should be passed only if you built a custom rollapp")
	cmd.Flags().StringP(FlagNames.TokenSupply, "", defaultTokenSupply, "The total token supply of the RollApp")
	cmd.Flags().BoolP(FlagNames.Interactive, "", false, "Run roller in interactive mode")
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
	cfg := utils.RollappConfig{}
	cfg.Home = initCmd.Flag(utils.FlagNames.Home).Value.String()
	cfg.RollappBinary = getRollappBinaryPath(initCmd)

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
	cfg.Denom = denom
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
