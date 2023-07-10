package initconfig

import (
	"fmt"
	"os"
	"strings"

	"github.com/dymensionxyz/roller/config"
	"github.com/manifoldco/promptui"
)

// TODO: return error output
func RunInteractiveMode(cfg *config.RollappConfig) {
	promptNetwork := promptui.Select{
		Label: "Select your network",
		Items: []string{"devnet", "local"},
	}
	_, mode, _ := promptNetwork.Run()
	cfg.HubData = Hubs[mode]

	promptChainID := promptui.Prompt{
		Label:     "Enter your RollApp ID",
		Default:   "myrollapp_1234-1",
		AllowEdit: true,
	}
	for {
		chainID, err := promptChainID.Run()
		if err != nil {
			break
		}
		err = config.ValidateRollAppID(chainID)
		if err != nil {
			fmt.Println(err)
			continue
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
	denom, _ := promptDenom.Run()
	cfg.Denom = "u" + denom

	promptTokenSupply := promptui.Prompt{
		Label:    "How many " + denom + " tokens do you wish to mint for Genesis?",
		Default:  "1000000000",
		Validate: config.VerifyTokenSupply,
	}
	supply, _ := promptTokenSupply.Run()
	cfg.TokenSupply = supply

	availableDAs := []config.DAType{config.Avail, config.Celestia}
	if mode == "local" {
		availableDAs = append(availableDAs, config.Mock)
	}
	promptDAType := promptui.Select{
		Label: "Choose your data layer",
		Items: availableDAs,
	}

	//TODO(#76): temporary hack to only support Celestia
	for {
		_, da, err := promptDAType.Run()
		if err != nil {
			break
		}
		da = strings.ToLower(da)
		if da == string(config.Avail) {
			fmt.Println("Avail not supported yet")
			continue
		}
		cfg.DA = config.DAType(da)
		break
	}

	promptExecutionEnv := promptui.Select{
		Label: "Choose your rollapp execution environment",
		Items: []string{"EVM rollapp", "custom"},
	}
	_, env, _ := promptExecutionEnv.Run()
	if env == "custom" {
		promptBinaryPath := promptui.Prompt{
			Label:     "Set your runtime binary",
			Default:   cfg.RollappBinary,
			AllowEdit: true,
			Validate: func(s string) error {
				_, err := os.Stat(s)
				return err
			},
		}
		path, _ := promptBinaryPath.Run()
		cfg.RollappBinary = path
	}
}
