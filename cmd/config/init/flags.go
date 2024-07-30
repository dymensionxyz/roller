package initconfig

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	globalutils "github.com/dymensionxyz/roller/utils"
)

func AddFlags(cmd *cobra.Command) error {
	cmd.Flags().
		StringP(
			FlagNames.HubID,
			"",
			consts.LocalHubName,
			fmt.Sprintf("The ID of the Dymension hub. %s", getAvailableHubsMessage()),
		)

	cmd.Flags().
		StringP(
			FlagNames.RollappBinary,
			"",
			consts.Executables.RollappEVM,
			"The rollapp binary. Should be passed only if you built a custom rollapp",
		)
	cmd.Flags().
		StringP(FlagNames.VMType, "", string(config.EVM_ROLLAPP), "The rollapp type [evm, sdk]. Defaults to evm")
	cmd.Flags().
		StringP(FlagNames.TokenSupply, "", consts.DefaultTokenSupply, "The total token supply of the RollApp")
	// cmd.Flags().BoolP(FlagNames.Interactive, "i", false, "Run roller in interactive mode")
	cmd.Flags().BoolP(FlagNames.NoOutput, "", false, "Run init without any output")
	cmd.Flags().UintP(
		FlagNames.Decimals, "", 18,
		"The precision level of the RollApp's token defined by the number of decimal places. "+
			"It should be an integer ranging between 1 and 18. This is akin to how 1 Ether equates to 10^18 Wei in Ethereum. "+
			"Note: EVM RollApps must set this value to 18.",
	)
	cmd.Flags().
		StringP(
			FlagNames.DAType,
			"",
			"Celestia",
			"The DA layer for the RollApp. Can be one of 'Celestia, Avail, Local'",
		)

	// TODO: Expose when supporting custom sdk rollapps.
	err := cmd.Flags().MarkHidden(FlagNames.Decimals)
	if err != nil {
		return err
	}
	return nil
}

func GetInitConfig(
	initCmd *cobra.Command,
	withMockSettlement bool,
) (*config.RollappConfig, error) {
	var cfg config.RollappConfig

	home, err := globalutils.ExpandHomePath(initCmd.Flag(utils.FlagNames.Home).Value.String())
	if err != nil {
		fmt.Println("failed to expand home path: ", err)
	}

	rollerConfigFilePath := filepath.Join(home, config.RollerConfigFileName)
	if _, err := toml.DecodeFile(rollerConfigFilePath, &cfg); err != nil {
		return nil, err
	}

	cfg.Home = home

	// TODO: support wasm, make the bainry name generic, like 'rollappd'
	// for both RollApp types
	cfg.RollappBinary = consts.Executables.RollappEVM

	// token supply is provided in the pre-created genesis
	// cfg.TokenSupply = initCmd.Flag(FlagNames.TokenSupply).Value.String()
	cfg.DA = config.DAType(strings.ToLower(string(cfg.DA)))

	var hubID string

	// TODO: hub id (and probably the rest of settlement config) should come from roller config
	if withMockSettlement {
		hubID = "mock"
	} else {
		hubID = initCmd.Flag(FlagNames.HubID).Value.String()
	}

	hub, ok := consts.Hubs[hubID]

	if !ok {
		return nil, fmt.Errorf("failed to retrieve the hub with hub id: %s", hubID)
	}

	cfg.HubData = hub

	// cfg.RollerVersion = version.TrimVersionStr(version.BuildVersion)
	// cfg.RollappID = raID
	// cfg.Denom = raBaseDenom

	if cfg.VMType == config.EVM_ROLLAPP {
		cfg.Decimals = 18
	} else {
		cfg.Decimals = 6
	}

	return &cfg, nil
}

func getAvailableHubsMessage() string {
	return fmt.Sprintf(
		"Acceptable values are '%s', '%s' or '%s'",
		consts.LocalHubName,
		consts.TestnetHubName,
		consts.MainnetHubName,
	)
}

// func isLowercaseAlphabetical(s string) bool {
// 	match, _ := regexp.MatchString("^[a-z]+$", s)
// 	return match
// }
