package initconfig

import (
	"fmt"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/utils/config"
	"github.com/dymensionxyz/roller/utils/filesystem"
)

type OutputHandler struct {
	*filesystem.OutputHandler
}

func NewOutputHandler(noOutput bool) *OutputHandler {
	return &OutputHandler{
		OutputHandler: filesystem.NewOutputHandler(noOutput),
	}
}

func (o *OutputHandler) PrintInitOutput(
	rollappConfig config.RollappConfig,
	addresses []utils.KeyInfo,
	rollappId string,
) {
	if o.NoOutput {
		return
	}
	fmt.Printf(
		"💈 RollApp '%s' configuration files have been successfully generated on your local machine. Congratulations!\n\n",
		rollappId,
	)

	if rollappConfig.HubData.ID == consts.MockHubID {
		PrintTokenSupplyLine(rollappConfig)
		fmt.Println()
	}
	utils.PrintAddressesWithTitle(addresses)

	if rollappConfig.HubData.ID != consts.MockHubID {
		pterm.DefaultSection.WithIndentCharacter("🔔").
			Println("Please fund the addresses below to register and run the rollapp.")
		fa := formatAddresses(rollappConfig, addresses)
		for _, v := range fa {
			v.Print(utils.WithName())
		}
	}
}

func (o *OutputHandler) PromptOverwriteConfig(home string) (bool, error) {
	if o.NoOutput {
		return true, nil
	}

	shouldOverwrite, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(
		fmt.Sprintf("Directory %s is not empty. Do you want to overwrite it?", home),
	).WithDefaultValue(false).Show()

	return shouldOverwrite, nil
}
