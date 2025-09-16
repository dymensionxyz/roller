package relayer

import (
	"fmt"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/config"
	"github.com/dymensionxyz/roller/utils/dependencies"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/roller"
)

// GetRollappToRunFor function retrieves the RollApp ID and Hub Data from the roller
// configuration file if it is present and returns
// the RollApp ID, Hub Data, keyring backend to use and error, if any.
// when no roller configuration file is present, it prompts the user for the
// necessary information and returns the RollApp ID, Hub Data, keyring backend to use
// and error, if any.
func GetRollappToRunFor(home, component string) (string, *consts.HubData, string, error) {
	rollerConfigFilePath := roller.GetConfigPath(home)
	rollerConfigExists, err := filesystem.DoesFileExist(rollerConfigFilePath)
	if err != nil {
		return "", nil, "", err
	}

	if rollerConfigExists {
		pterm.Info.Println(
			"existing roller configuration found, retrieving RollApp ID from it",
		)
		rollerData, err := roller.LoadConfig(home)
		if err != nil {
			pterm.Error.Printf("failed to load rollapp config: %v\n", err)
			return "", nil, "", err
		}

		msg := fmt.Sprintf(
			"the retrieved rollapp ID is: %s, would you like to initialize the %s for this rollapp?",
			pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
				Sprint(rollerData.RollappID),
			component,
		)
		runForRollappFromRollerConfig, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(msg).
			Show()

		if runForRollappFromRollerConfig {
			raID := rollerData.RollappID

			rollerData = config.PromptCustomHubEndpoint(rollerData)
			hd := rollerData.HubData

			return raID, &hd, string(rollerData.KeyringBackend), nil
		}
	}

	return promptForRaAndHd()
}

func NewIbcConnenctionCanBeCreatedOnCurrentNode(home, raID string) (bool, error) {
	rollerData, err := roller.LoadConfig(home)
	if err != nil {
		pterm.Error.Printf("failed to load rollapp config: %v\n", err)
		return false, err
	}

	err = fmt.Errorf(
		"new channels can only be initialized on a sequencer node which is running for a rollapp you're trying to create the IBC connection for",
	)

	runForRollappFromRollerConfig := raID == rollerData.RollappID
	if !runForRollappFromRollerConfig {
		return false, err
	}

	if runForRollappFromRollerConfig && rollerData.NodeType != consts.NodeType.Sequencer {
		return false, err
	}

	return true, nil
}

// promptForRaAndHd function prompts the user for the RollApp ID and Hub Data
// and returns the RollApp ID, Hub Data, keyring backend to use and error, if any
func promptForRaAndHd() (string, *consts.HubData, string, error) {
	var hd consts.HubData

	raID := config.PromptRaID()
	env := config.PromptEnvironment()

	if env == "playground" || env == "blumbus" || env == "mainnet" {
		hd = consts.Hubs[env]
	} else {
		chd, err := config.CreateCustomHubData("")

		hd = consts.HubData{
			Environment:   env,
			ID:            chd.ID,
			ApiUrl:        chd.ApiUrl,
			RpcUrl:        chd.RpcUrl,
			ArchiveRpcUrl: chd.RpcUrl,
			GasPrice:      "2000000000",
		}
		if err != nil {
			return "", nil, "", err
		}

		err = dependencies.InstallCustomDymdVersion(chd.DymensionHash)
		if err != nil {
			pterm.Error.Println("failed to install custom dymd version: ", err)
			return "", nil, "", err
		}
	}

	return raID, &hd, "test", nil
}
