package initconfig

import (
	"fmt"

	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/manifoldco/promptui"
)

func RunInteractiveMode(config *utils.RollappConfig) {
	promptChainID := promptui.Prompt{
		Label:    "Chain ID",
		Validate: utils.ValidateRollAppID,
	}
	chainID, _ := promptChainID.Run()
	config.RollappID = chainID

	// TODO: add validation for denomination
	promptDenom := promptui.Prompt{
		Label: "Denomination",
	}
	denom, _ := promptDenom.Run()
	config.Denom = denom

	promptTokenSupply := promptui.Prompt{
		Label:    "TokenSupply",
		Validate: utils.VerifyTokenSupply,
	}
	supply, _ := promptTokenSupply.Run()
	config.TokenSupply = supply

	promptType := promptui.Select{
		Label: "CLI type",
		Items: []string{"evm", "sdk"},
	}
	_, _, _ = promptType.Run()
	fmt.Println("Only EVM supported for now")

	promptDAType := promptui.Select{
		Label: "DA type",
		Items: []string{"Celestia", "Avail"},
	}
	_, _, _ = promptDAType.Run()
	fmt.Println("Only Celestia supported for now")

	promptNetwork := promptui.Select{
		Label: "Network",
		Items: []string{"local", "internal-devnet"},
	}
	_, mode, _ := promptNetwork.Run()
	config.HubData = Hubs[mode]
}
