package initconfig

import (
	"fmt"

	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/manifoldco/promptui"
)

func RunInteractiveMode(config *utils.RollappConfig) {
	promptChainID := promptui.Prompt{
		Label: "Chain ID",
	}
	chainID, _ := promptChainID.Run()
	config.RollappID = chainID

	promptDenom := promptui.Prompt{
		Label: "Denomination",
	}
	denom, _ := promptDenom.Run()
	config.Denom = denom

	promptTokenSupply := promptui.Prompt{
		Label: "TokenSupply",
	}
	supply, _ := promptTokenSupply.Run()
	config.TokenSupply = supply

	promptType := promptui.Select{
		Label: "CLI type",
		Items: []string{"evm", "sdk"},
	}
	_, _, _ = promptType.Run()
	// config.CliType = cliType
	fmt.Println("Only EVM supported for now")

	promptDAType := promptui.Select{
		Label: "DA type",
		Items: []string{"Celestia", "Avail"},
	}
	_, _, _ = promptDAType.Run()
	// config.daType = daType
	fmt.Println("Only Celestia supported for now")

	promptNetwork := promptui.Select{
		Label: "Network",
		Items: []string{"local", "internal-devnet"},
	}
	_, mode, _ := promptNetwork.Run()
	//TODO: add validation
	config.HubData = Hubs[mode]

	// promptPrefix := promptui.Prompt{
	// 	Label:    "Prefix",
	// 	Default:  config.Denom,
	// 	Validate: validateNotEmpty,
	// }
	// prefix, _ := promptPrefix.Run()
	// config.Prefix = prefix

	// promptGenesisAccountAddr := promptui.Prompt{
	// 	Label: "Genesis account address",
	// }
	// genesisAccountAddr, _ := promptGenesisAccountAddr.Run()
	// if genesisAccountAddr != "" {
	// 	config.GenesisAccountAddr = genesisAccountAddr
	// 	config.NoValidators = true
	// }
}

// func validateNotEmpty(input string) error {
// 	if input == "" {
// 		return fmt.Errorf("input must not be empty")
// 	}
// 	return nil
// }
