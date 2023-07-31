package initconfig

import (
	"fmt"
	"os"
	"strings"

	"github.com/dymensionxyz/roller/config"
	"github.com/manifoldco/promptui"
)

func RunInteractiveMode(cfg *config.RollappConfig) error {
	promptNetwork := promptui.Select{
		Label: "Select your network",
		Items: []string{"devnet", "local"},
	}
	_, mode, err := promptNetwork.Run()
	if err != nil {
		return err
	}
	cfg.HubData = Hubs[mode]
	cfg.VMType = config.EVM_ROLLAPP
	promptExecutionEnv := promptui.Select{
		Label: "Choose your rollapp execution environment",
		Items: []string{"EVM rollapp", "custom EVM rollapp", "custom non-EVM rollapp"},
	}
	_, env, err := promptExecutionEnv.Run()
	if err != nil {
		return err
	}

	if env != "EVM rollapp" {
		if env == "custom non-EVM rollapp" {
			cfg.VMType = config.SDK_ROLLAPP
		}
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
		Default:   "myrollapp_1234-1",
		AllowEdit: true,
	}
	for {
		chainID, err := promptChainID.Run()
		if err != nil {
			return err
		}
		if cfg.VMType == config.EVM_ROLLAPP {
			err = config.ValidateRollAppID(chainID)
			if err != nil {
				fmt.Println(err)
				continue
			}
		}
		err = VerifyUniqueRollappID(chainID, *cfg)
		if err != nil {
			fmt.Println(err)
			continue
		}
		cfg.RollappID = chainID
		break
	}

	promptDenom := promptui.Prompt{
		Label:     "Specify your RollApp denom",
		Default:   "RAX",
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
	cfg.Denom = "u" + denom

	promptTokenSupply := promptui.Prompt{
		Label:    "How many " + denom + " tokens do you wish to mint for Genesis?",
		Default:  "1000000000",
		Validate: config.VerifyTokenSupply,
	}
	supply, err := promptTokenSupply.Run()
	if err != nil {
		return err
	}
	cfg.TokenSupply = supply

	availableDAs := []config.DAType{config.Celestia, config.Avail}
	if mode == "local" {
		availableDAs = append(availableDAs, config.Mock)
	}
	promptDAType := promptui.Select{
		Label: "Choose your data layer",
		Items: availableDAs,
	}
	_, da, _ := promptDAType.Run()
	cfg.DA = config.DAType(strings.ToLower(da))

	return nil
}
