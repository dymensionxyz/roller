package initconfig

import (
	"fmt"

	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/manifoldco/promptui"
)

// TODO: return error output
func RunInteractiveMode(config *utils.RollappConfig) {
	promptNetwork := promptui.Select{
		Label:     "Select your network",
		Items:     []string{"local", "devnet"},
		CursorPos: 1,
	}
	_, mode, _ := promptNetwork.Run()
	config.HubData = Hubs[mode]

	promptChainID := promptui.Prompt{
		Label:     "Enter your RollApp ID",
		Default:   "myrollapp_1234-1",
		AllowEdit: true,
		Validate:  utils.ValidateRollAppID,
	}
	chainID, _ := promptChainID.Run()
	config.RollappID = chainID

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
	_, _, _ = promptDAType.Run()
	fmt.Println("Only Celestia supported for now")

	promptBinaryPath := promptui.Prompt{
		Label:     "Set your runtime binary",
		Default:   config.RollappBinary,
		AllowEdit: true,
		//TODO: add validate for binary path
	}
	path, _ := promptBinaryPath.Run()
	config.RollappBinary = path
}
