package initconfig

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	global_utils "github.com/dymensionxyz/roller/utils"
	"github.com/dymensionxyz/roller/version"
)

const (
	defaultTokenSupply = "1000000000"
)

func AddFlags(cmd *cobra.Command) error {
	cmd.Flags().
		StringP(FlagNames.HubID, "", consts.LocalHubName, fmt.Sprintf("The ID of the Dymension hub. %s", getAvailableHubsMessage()))
	cmd.Flags().
		StringP(FlagNames.RollappBinary, "", consts.Executables.RollappEVM, "The rollapp binary. Should be passed only if you built a custom rollapp")
	cmd.Flags().
		StringP(FlagNames.VMType, "", string(config.EVM_ROLLAPP), "The rollapp type [evm, sdk]. Defaults to evm")
	cmd.Flags().
		StringP(FlagNames.TokenSupply, "", defaultTokenSupply, "The total token supply of the RollApp")
	// cmd.Flags().BoolP(FlagNames.Interactive, "i", false, "Run roller in interactive mode")
	cmd.Flags().BoolP(FlagNames.NoOutput, "", false, "Run init without any output")
	cmd.Flags().UintP(FlagNames.Decimals, "", 18,
		"The precision level of the RollApp's token defined by the number of decimal places. "+
			"It should be an integer ranging between 1 and 18. This is akin to how 1 Ether equates to 10^18 Wei in Ethereum. "+
			"Note: EVM RollApps must set this value to 18.")
	cmd.Flags().
		StringP(FlagNames.DAType, "", "Celestia", "The DA layer for the RollApp. Can be one of 'Celestia, Avail, Local'")

	// TODO: Expose when supporting custom sdk rollapps.
	err := cmd.Flags().MarkHidden(FlagNames.Decimals)
	if err != nil {
		return err
	}
	return nil
}

func GetInitConfig(initCmd *cobra.Command, args []string) (*config.RollappConfig, error) {
	home := initCmd.Flag(utils.FlagNames.Home).Value.String()
	rollerConfigFilePath := filepath.Join(home, "roller.toml")
	// interactive, _ := initCmd.Flags().GetBool(FlagNames.Interactive)
	raType, err := global_utils.GetKeyFromTomlFile(rollerConfigFilePath, "execution")
	if err != nil {
		return nil, err
	}
	raID, err := global_utils.GetKeyFromTomlFile(rollerConfigFilePath, "rollapp_id")
	if err != nil {
		return nil, err
	}
	raBaseDenom, err := global_utils.GetKeyFromTomlFile(rollerConfigFilePath, "base_denom")
	if err != nil {
		return nil, err
	}
	da, err := global_utils.GetKeyFromTomlFile(rollerConfigFilePath, "da")
	if err != nil {
		return nil, err
	}

	fmt.Println(home, raType, raID, raBaseDenom, da)

	// load initial config if exists
	var cfg config.RollappConfig
	// load from flags
	cfg.Home = home

	// TODO: support wasm, make the bainry name generic, like 'rollappd'
	// for both RollApp types
	cfg.RollappBinary = consts.Executables.RollappEVM
	cfg.VMType = config.VMType(raType)
	// token supply is provided in the pre-created genesis
	// cfg.TokenSupply = initCmd.Flag(FlagNames.TokenSupply).Value.String()
	// decimals, _ := initCmd.Flags().GetUint(FlagNames.Decimals)
	cfg.Decimals = 18
	cfg.DA = config.DAType(strings.ToLower(da))

	hubID := initCmd.Flag(FlagNames.HubID).Value.String()
	if hub, ok := consts.Hubs[hubID]; ok {
		cfg.HubData = hub
	}
	cfg.RollerVersion = version.TrimVersionStr(version.BuildVersion)
	cfg.RollappID = raID
	cfg.Denom = raBaseDenom

	return formatBaseCfg(cfg, initCmd)
}

func formatBaseCfg(
	cfg config.RollappConfig,
	initCmd *cobra.Command,
) (*config.RollappConfig, error) {
	setDecimals(initCmd, &cfg)
	return &cfg, nil
}

// there's no need for a custom chain id as the rollapp id is set via the genesis-creator
// func generateRollappId(rlpCfg config.RollappConfig) (string, error) {
// 	for {
// 		RandEthId, err := generateRandEthId()
// 		if err != nil {
// 			return "", err
// 		}
// 		if rlpCfg.HubData.ID == consts.LocalHubID {
// 			return fmt.Sprintf("%s_%s-1", rlpCfg.RollappID, RandEthId), nil
// 		}
// 		isUnique, err := isEthIdentifierUnique(RandEthId, rlpCfg)
// 		if err != nil {
// 			return "", err
// 		}
// 		if isUnique {
// 			return fmt.Sprintf("%s_%s-1", rlpCfg.RollappID, RandEthId), nil
// 		}
// 	}
// }

// func generateRandEthId() (string, error) {
// 	max := big.NewInt(9000000)
// 	n, err := rand.Int(rand.Reader, max)
// 	if err != nil {
// 		return "", err
// 	}
// 	return fmt.Sprintf("%d", n), nil
// }

func setDecimals(initCmd *cobra.Command, cfg *config.RollappConfig) {
	decimals, _ := initCmd.Flags().GetUint(FlagNames.Decimals)
	if cfg.VMType == config.EVM_ROLLAPP || initCmd.Flags().Lookup(FlagNames.Decimals).Changed {
		cfg.Decimals = decimals
	} else {
		cfg.Decimals = 6
	}
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
