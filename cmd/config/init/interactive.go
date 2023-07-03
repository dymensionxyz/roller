package initconfig

import (
	"fmt"
	"os"

	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/manifoldco/promptui"
)

// TODO: return error output
func RunInteractiveMode(config *utils.RollappConfig) {
	promptNetwork := promptui.Select{
		Label: "Select your network",
		Items: []string{"devnet", "local"},
	}
	_, mode, _ := promptNetwork.Run()
	config.HubData = Hubs[mode]

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
		if err := utils.ValidateRollAppID(chainID); err == nil {
			config.RollappID = chainID
			break
		}
		fmt.Println("Expected format: name_uniqueID-revision (e.g. myrollapp_1234-1)")
	}

	promptDenom := promptui.Prompt{
		Label:   "Specify your RollApp denom",
		Default: "RAX",
		// TODO: add validation for denomination
	}
	denom, _ := promptDenom.Run()
	config.Denom = denom

	promptTokenSupply := promptui.Prompt{
		Label:    "How many " + denom + " tokens do you wish to mint for Genesis?",
		Default:  "1000000000",
		Validate: utils.VerifyTokenSupply,
	}
	supply, _ := promptTokenSupply.Run()
	config.TokenSupply = supply

	promptDAType := promptui.Select{
		Label: "Choose your data layer",
		Items: []string{"Celestia", "Avail"},
	}

	//TODO(#76): temporary hack to only support Celestia
	for {
		_, da, err := promptDAType.Run()
		if err != nil || da == "Celestia" {
			break
		}
		if da != "Celestia" {
			fmt.Println("Only Celestia supported for now")
		}
	}

	promptExecutionEnv := promptui.Select{
		Label: "Choose your rollapp execution environment",
		Items: []string{"EVM rollapp", "custom"},
	}
	_, env, _ := promptExecutionEnv.Run()
	if env == "SDK" {
		promptBinaryPath := promptui.Prompt{
			Label:     "Set your runtime binary",
			Default:   config.RollappBinary,
			AllowEdit: true,
			Validate: func(s string) error {
				_, err := os.Stat(s)
				return err
			},
		}
		path, _ := promptBinaryPath.Run()
		config.RollappBinary = path
	}
}
