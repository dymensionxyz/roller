package initconfig

import (
	"fmt"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/utils"
)

func addFlags(cmd *cobra.Command) {
	cmd.Flags().StringP(FlagNames.HubID, "", TestnetHubID, fmt.Sprintf("The ID of the Dymension hub. %s", getAvailableHubsMessage()))
	cmd.Flags().StringP(FlagNames.DAEndpoint, "", "", "The data availability light node endpoint. Runs an Arabica Celestia light node if not provided")
	cmd.Flags().StringP(FlagNames.RollappBinary, "", "", "The rollapp binary. Should be passed only if you built a custom rollapp")
	cmd.Flags().Uint64P(FlagNames.Decimals, "", 18, "The number of decimal places a rollapp token supports")
	cmd.Flags().StringP(FlagNames.Home, "", utils.GetRollerRootDir(), "The directory of the roller config files")

	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		hubID, err := cmd.Flags().GetString(FlagNames.HubID)
		if err != nil {
			return err
		}
		if _, ok := Hubs[hubID]; !ok {
			return fmt.Errorf("invalid hub ID: %s. %s", hubID, getAvailableHubsMessage())
		}
		return nil
	}
}

func getDecimals(cmd *cobra.Command) uint64 {
	decimals, err := cmd.Flags().GetUint64(FlagNames.Decimals)
	if err != nil {
		panic(err)
	}
	return decimals
}

func getRollappBinaryPath(cmd *cobra.Command) string {
	rollappBinaryPath := cmd.Flag(FlagNames.RollappBinary).Value.String()
	if rollappBinaryPath == "" {
		return consts.Executables.RollappEVM
	}
	return rollappBinaryPath
}

func GetInitConfig(cmd *cobra.Command, args []string) InitConfig {
	rollappId := args[0]
	denom := args[1]
	home := cmd.Flag(FlagNames.Home).Value.String()
	createLightNode := !cmd.Flags().Changed(FlagNames.DAEndpoint)
	rollappBinaryPath := getRollappBinaryPath(cmd)
	decimals := getDecimals(cmd)
	hubID := cmd.Flag(FlagNames.HubID).Value.String()
	return InitConfig{
		Home:              home,
		RollappID:         rollappId,
		RollappBinary:     rollappBinaryPath,
		CreateDALightNode: createLightNode,
		Denom:             denom,
		Decimals:          decimals,
		HubData:           Hubs[hubID],
	}
}

func getAvailableHubsMessage() string {
	return fmt.Sprintf("Acceptable values are '%s', '%s' or '%s'", TestnetHubID, StagingHubID, LocalHubID)
}
