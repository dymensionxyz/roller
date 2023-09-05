package initconfig

import (
	"fmt"
	"math/rand"
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
	cmd.Flags().StringP(FlagNames.HubID, "", FroopylandHubName, fmt.Sprintf("The ID of the Dymension hub. %s", getAvailableHubsMessage()))
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

func GetInitConfig(initCmd *cobra.Command, args []string) (config.RollappConfig, error) {
	cfg := config.RollappConfig{
		RollerVersion: version.BuildVersion,
	}
	cfg.Home = initCmd.Flag(utils.FlagNames.Home).Value.String()
	cfg.RollappBinary = initCmd.Flag(FlagNames.RollappBinary).Value.String()
	// Error is ignored because the flag is validated in the cobra preRun hook
	interactive, _ := initCmd.Flags().GetBool(FlagNames.Interactive)
	if interactive {
		if err := RunInteractiveMode(&cfg); err != nil {
			return cfg, err
		}
		return formatBaseCfg(cfg, initCmd), nil
	}

	rollappId := args[0]
	denom := args[1]
	if !isAlphanumeric(rollappId) {
		return cfg, fmt.Errorf("invalid rollapp id %s. %s", rollappId, validRollappIDMsg)
	}
	hubID := initCmd.Flag(FlagNames.HubID).Value.String()
	tokenSupply := initCmd.Flag(FlagNames.TokenSupply).Value.String()
	cfg.RollappID = rollappId
	cfg.Denom = "u" + denom
	cfg.HubData = Hubs[hubID]
	cfg.TokenSupply = tokenSupply
	cfg.DA = config.DAType(strings.ToLower(initCmd.Flag(FlagNames.DAType).Value.String()))
	cfg.VMType = config.VMType(initCmd.Flag(FlagNames.VMType).Value.String())
	return formatBaseCfg(cfg, initCmd), nil
}

func formatBaseCfg(cfg config.RollappConfig, initCmd *cobra.Command) config.RollappConfig {
	setDecimals(initCmd, &cfg)
	formattedRollappId, err := generateRollappId(cfg)
	if err != nil {
		return cfg
	}
	cfg.RollappID = formattedRollappId
	return cfg
}

func generateRollappId(rlpCfg config.RollappConfig) (string, error) {
	for {
		RandEthId := generateRandEthId()
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
	return fmt.Sprintf("Acceptable values are '%s', '%s' or '%s'", FroopylandHubName, StagingHubName, LocalHubName)
}
