package initconfig

import (
	"fmt"
	"os"
	"strings"

	"github.com/manifoldco/promptui"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/config"
)

func RunInteractiveMode(cfg *config.RollappConfig) error {
	promptNetwork := promptui.Select{
		Label: "Select your network",
		Items: []string{"froopyland", "devnet", "local"},
		CursorPos: func() int {
			switch cfg.HubData.ID {
			case consts.Hubs[consts.FroopylandHubName].ID:
				return 0
			case consts.Hubs[consts.StagingHubName].ID:
				return 1
			case consts.Hubs[consts.LocalHubName].ID:
				return 2
			default:
				return 0
			}
		}(),
	}
	_, mode, err := promptNetwork.Run()
	if err != nil {
		return err
	}
	cfg.HubData = consts.Hubs[mode]

	promptExecutionEnv := promptui.Select{
		Label: "Choose your rollapp execution environment",
		Items: []string{"EVM rollapp", "custom EVM rollapp", "custom non-EVM rollapp"},
		CursorPos: func() int {
			if cfg.RollappBinary == consts.Executables.RollappEVM {
				return 0
			} else if cfg.VMType == config.EVM_ROLLAPP {
				return 1
			} else {
				return 2
			}
		}(),
	}
	_, env, err := promptExecutionEnv.Run()
	if err != nil {
		return err
	}

	if env == "custom non-EVM rollapp" {
		cfg.VMType = config.SDK_ROLLAPP
	} else {
		cfg.VMType = config.EVM_ROLLAPP
	}

	// if custom binary, get the binary path
	if env != "EVM rollapp" {
		promptBinaryPath := promptui.Prompt{
			Label:     "Set your runtime binary",
			Default:   cfg.RollappBinary,
			AllowEdit: true,
			Validate: func(s string) error {
				_, err := os.Stat(s)
				return err
			},
		}
		path, err := promptBinaryPath.Run()
		if err != nil {
			return err
		}
		cfg.RollappBinary = path
	}
	promptChainID := promptui.Prompt{
		Label:     "Enter your RollApp ID",
		Default:   strings.Split(cfg.RollappID, "_")[0],
		AllowEdit: true,
	}
	for {
		rollappID, err := promptChainID.Run()
		if err != nil {
			return err
		}
		if !isLowercaseAlphabetical(rollappID) {
			fmt.Printf("invalid rollapp id %s. %s\n", rollappID, validRollappIDMsg)
			continue
		}
		cfg.RollappID = rollappID
		break
	}

	promptDenom := promptui.Prompt{
		Label:     "Specify your RollApp denom",
		Default:   strings.TrimPrefix(cfg.Denom, "a"),
		AllowEdit: true,
		Validate: func(s string) error {
			if !config.IsValidTokenSymbol(s) {
				return fmt.Errorf("invalid token symbol")
			}
			return nil
		},
	}
	denom, err := promptDenom.Run()
	if err != nil {
		return err
	}
	cfg.Denom = "a" + denom

	promptTokenSupply := promptui.Prompt{
		Label:    "How many " + denom + " tokens do you wish to mint for Genesis?",
		Default:  cfg.TokenSupply,
		Validate: config.VerifyTokenSupply,
	}
	supply, err := promptTokenSupply.Run()
	if err != nil {
		return err
	}
	cfg.TokenSupply = supply

	availableDAs := []config.DAType{config.Celestia, config.Avail, config.Local}
	promptDAType := promptui.Select{
		Label: "Choose your data layer",
		Items: availableDAs,
		CursorPos: func() int {
			switch cfg.DA {
			case config.Celestia:
				return 0
			case config.Avail:
				return 1
			case config.Local:
				return 2
			default:
				return 0
			}
		}(),
	}
	_, da, _ := promptDAType.Run()
	cfg.DA = config.DAType(strings.ToLower(da))

	return nil
}
