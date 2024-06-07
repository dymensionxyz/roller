package initconfig

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/dymensionxyz/roller/version"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"github.com/spf13/cobra"
)

const (
	defaultTokenSupply = "1000000000"
)

func addFlags(cmd *cobra.Command) error {
	cmd.Flags().StringP(FlagNames.HubID, "", consts.FroopylandHubName, fmt.Sprintf("The ID of the Dymension hub. %s", getAvailableHubsMessage()))
	cmd.Flags().StringP(FlagNames.RollappBinary, "", consts.Executables.RollappEVM, "The rollapp binary. Should be passed only if you built a custom rollapp")
	cmd.Flags().StringP(FlagNames.VMType, "", string(config.EVM_ROLLAPP), "The rollapp type [evm, sdk]. Defaults to evm")
	cmd.Flags().StringP(FlagNames.TokenSupply, "", defaultTokenSupply, "The total token supply of the RollApp")
	cmd.Flags().BoolP(FlagNames.Interactive, "i", false, "Run roller in interactive mode")
	cmd.Flags().BoolP(FlagNames.NoOutput, "", false, "Run init without any output")
	cmd.Flags().UintP(FlagNames.Decimals, "", 18,
		"The precision level of the RollApp's token defined by the number of decimal places. "+
			"It should be an integer ranging between 1 and 18. This is akin to how 1 Ether equates to 10^18 Wei in Ethereum. "+
			"Note: EVM RollApps must set this value to 18.")
	cmd.Flags().StringP(FlagNames.DAType, "", "Celestia", "The DA layer for the RollApp. Can be one of 'Celestia, Avail, Local'")

	// TODO: Expose when supporting custom sdk rollapps.
	err := cmd.Flags().MarkHidden(FlagNames.Decimals)
	if err != nil {
		return err
	}
	return nil
}

func GetInitConfig(initCmd *cobra.Command, args []string) (*config.RollappConfig, error) {
	home := initCmd.Flag(utils.FlagNames.Home).Value.String()
	interactive, _ := initCmd.Flags().GetBool(FlagNames.Interactive)

	// load initial config if exists
	var cfg config.RollappConfig
	// load from flags
	cfg.Home = home
	cfg.RollappBinary = initCmd.Flag(FlagNames.RollappBinary).Value.String()
	cfg.VMType = config.VMType(initCmd.Flag(FlagNames.VMType).Value.String())
	cfg.TokenSupply = initCmd.Flag(FlagNames.TokenSupply).Value.String()
	decimals, _ := initCmd.Flags().GetUint(FlagNames.Decimals)
	cfg.Decimals = decimals
	cfg.DA = config.DAType(strings.ToLower(initCmd.Flag(FlagNames.DAType).Value.String()))
	hubID := initCmd.Flag(FlagNames.HubID).Value.String()
	if hub, ok := consts.Hubs[hubID]; ok {
		cfg.HubData = hub
	}
	cfg.RollerVersion = version.TrimVersionStr(version.BuildVersion)

	if len(args) > 0 {
		cfg.RollappID = args[0]
	}
	if len(args) > 1 {
		cfg.Denom = "a" + args[1]
	}

	if interactive {
		if err := RunInteractiveMode(&cfg); err != nil {
			return nil, err
		}
	}

	return formatBaseCfg(cfg, initCmd)
}

func formatBaseCfg(cfg config.RollappConfig, initCmd *cobra.Command) (*config.RollappConfig, error) {
	setDecimals(initCmd, &cfg)
	formattedRollappId, err := generateRollappId(cfg)
	if err != nil {
		return nil, err
	}
	cfg.RollappID = formattedRollappId
	return &cfg, nil
}

func generateRollappId(rlpCfg config.RollappConfig) (string, error) {
	for {
		RandEthId := generateRandEthId()
		if rlpCfg.HubData.ID == consts.LocalHubID {
			return fmt.Sprintf("%s_%s-1", rlpCfg.RollappID, RandEthId), nil
		}
		isUnique, err := isEthIdentifierUnique(RandEthId, rlpCfg)
		if err != nil {
			return "", err
		}
		if isUnique {
			return fmt.Sprintf("%s_%s-1", rlpCfg.RollappID, RandEthId), nil
		}
	}
}

func generateRandEthId() string {
	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Intn(9000000) + 1000000
	return fmt.Sprintf("%d", randomNumber)
}

func setDecimals(initCmd *cobra.Command, cfg *config.RollappConfig) {
	decimals, _ := initCmd.Flags().GetUint(FlagNames.Decimals)
	if cfg.VMType == config.EVM_ROLLAPP || initCmd.Flags().Lookup(FlagNames.Decimals).Changed {
		cfg.Decimals = decimals
	} else {
		cfg.Decimals = 6
	}
}

func getAvailableHubsMessage() string {
	return fmt.Sprintf("Acceptable values are '%s', '%s' or '%s'", consts.FroopylandHubName, consts.StagingHubName, consts.LocalHubName)
}

func isLowercaseAlphabetical(s string) bool {
	match, _ := regexp.MatchString("^[a-z]+$", s)
	return match
}
