package initconfig

import (
	"github.com/spf13/cobra"
)

func addFlags(cmd *cobra.Command) {
	cmd.Flags().StringP(FlagNames.HubRPC, "", HubData.RPC_URL, "Dymension Hub rpc endpoint")
	cmd.Flags().StringP(FlagNames.DAEndpoint, "", "", "The data availability light node endpoint. Runs an Arabica Celestia light node if not provided")
	cmd.Flags().StringP(FlagNames.RollappBinary, "", "", "The rollapp binary. Should be passed only if you built a custom rollapp")
	cmd.Flags().Uint64P(FlagNames.Decimals, "", 18, "The number of decimal places a rollapp token supports")
	cmd.Flags().StringP(FlagNames.Home, "", getRollerRootDir(), "The directory of the roller config files")
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
		return defaultRollappBinaryPath
	}
	return rollappBinaryPath
}

func getInitConfig(cmd *cobra.Command, args []string) InitConfig {
	rollappId := args[0]
	denom := args[1]
	home := cmd.Flag(FlagNames.Home).Value.String()
	createLightNode := !cmd.Flags().Changed(FlagNames.DAEndpoint)
	rollappBinaryPath := getRollappBinaryPath(cmd)
	decimals := getDecimals(cmd)
	return InitConfig{
		Home:              home,
		RollappID:         rollappId,
		RollappBinary:     rollappBinaryPath,
		CreateDALightNode: createLightNode,
		Denom:             denom,
		HubID:             HubData.ID,
		Decimals:          decimals,
	}
}
