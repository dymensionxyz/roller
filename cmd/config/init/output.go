package initconfig

import (
	"fmt"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"github.com/manifoldco/promptui"
)

type OutputHandler struct {
	*utils.OutputHandler
}

func NewOutputHandler(noOutput bool) *OutputHandler {
	return &OutputHandler{
		OutputHandler: utils.NewOutputHandler(noOutput),
	}
}

func (o *OutputHandler) printInitOutput(rollappConfig config.RollappConfig, addresses []utils.AddressData, rollappId string) {
	if o.NoOutput {
		return
	}
	fmt.Printf("💈 RollApp '%s' configuration files have been successfully generated on your local machine. Congratulations!\n\n", rollappId)
	fmt.Println(FormatTokenSupplyLine(rollappConfig))
	fmt.Println()
	utils.PrintAddresses(formatAddresses(rollappConfig, addresses))
	fmt.Printf("\n🔔 Please fund these addresses to register and run the rollapp.\n")
}

func (o *OutputHandler) PromptOverwriteConfig(home string) (bool, error) {
	if o.NoOutput {
		return true, nil
	}
	prompt := promptui.Prompt{
		Label:     fmt.Sprintf("Directory %s is not empty. Do you want to overwrite", home),
		IsConfirm: true,
	}
	_, err := prompt.Run()
	if err != nil {
		if err == promptui.ErrAbort {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
