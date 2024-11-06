package relayer

import (
	"fmt"
	"path/filepath"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/utils/config"
	"github.com/dymensionxyz/roller/utils/dependencies"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/roller"
)

func GetRollappToRunFor(home string) (string, *consts.HubData, error) {
	rollerConfigFilePath := roller.GetConfigPath(home)
	rollerConfigExists, err := filesystem.DoesFileExist(rollerConfigFilePath)
	if err != nil {
		return "", nil, err
	}

	if rollerConfigExists {
		pterm.Info.Println(
			"existing roller configuration found, retrieving RollApp ID from it",
		)
		rollerData, err := roller.LoadConfig(home)
		if err != nil {
			pterm.Error.Printf("failed to load rollapp config: %v\n", err)
			return "", nil, err
		}

		msg := fmt.Sprintf(
			"the retrieved rollapp ID is: %s, would you like to initialize the relayer for this rollapp?",
			pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
				Sprint(rollerData.RollappID),
		)
		runForRollappFromRollerConfig, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(msg).
			Show()

		if runForRollappFromRollerConfig {
			raID := rollerData.RollappID
			hd := rollerData.HubData

			return raID, &hd, nil
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

func promptForRaAndHd() (string, *consts.HubData, error) {
	var hd consts.HubData

	raID := config.PromptRaID()
	env := config.PromptEnvironment()

	if env == "playground" {
		hd = consts.Hubs[env]
	} else {
		chd, err := config.CreateCustomHubData()
		hd = *chd
		if err != nil {
			return "", nil, err
		}

		err = dependencies.InstallCustomDymdVersion()
		if err != nil {
			pterm.Error.Println("failed to install custom dymd version: ", err)
			return "", nil, err
		}
	}

	return raID, &hd, nil
}

func VerifyRelayerBalances(hd consts.HubData) error {
	insufficientBalances, err := getRelayerInsufficientBalances(hd)
	if err != nil {
		return err
	}

	if len(insufficientBalances) != 0 {
		err = keys.PrintInsufficientBalancesIfAny(insufficientBalances)
		if err != nil {
			return err
		}
	}

	return nil
}

func getRelayerInsufficientBalances(
	hd consts.HubData,
) ([]keys.NotFundedAddressData, error) {
	var insufficientBalances []keys.NotFundedAddressData
	home, err := roller.GetRootDir()
	if err != nil {
		return nil, err
	}

	accData, err := GetRelayerAccountsData(home, hd)
	if err != nil {
		return nil, err
	}

	// consts.Denoms.Hub is used here because as of @202409 we no longer require rollapp
	// relayer account funding to establish IBC connection.
	for _, acc := range accData {
		if acc.Balance.Amount.Cmp(oneDayRelayPrice.BigInt()) < 0 {
			insufficientBalances = append(
				insufficientBalances, keys.NotFundedAddressData{
					KeyName:         consts.KeysIds.HubRelayer,
					Address:         acc.Address,
					CurrentBalance:  acc.Balance.Amount,
					RequiredBalance: oneDayRelayPrice.BigInt(),
					Denom:           consts.Denoms.Hub,
					Network:         hd.ID,
				},
			)
		}
	}

	return insufficientBalances, nil
}

func GetRelayerAccountsData(
	home string,
	hd consts.HubData,
) ([]keys.AccountData, error) {
	var data []keys.AccountData

	// rollappRlyAcc, err := getRolRlyAccData(cfg)
	// if err != nil {
	// 	return nil, err
	// }
	// data = append(data, *rollappRlyAcc)

	hubRlyAcc, err := getHubRlyAccData(home, hd)
	if err != nil {
		return nil, err
	}

	data = append(data, *hubRlyAcc)
	return data, nil
}

// nolint: unused
func getRolRlyAccData(home string, raData roller.RollappConfig) (*keys.AccountData, error) {
	RollappRlyAddr, err := keys.GetRelayerAddress(home, raData.RollappID)
	seq := sequencer.GetInstance(raData)
	if err != nil {
		return nil, err
	}

	RollappRlyBalance, err := keys.QueryBalance(
		keys.ChainQueryConfig{
			RPC:    seq.GetRPCEndpoint(),
			Denom:  raData.Denom,
			Binary: consts.Executables.RollappEVM,
		}, RollappRlyAddr,
	)
	if err != nil {
		return nil, err
	}

	return &keys.AccountData{
		Address: RollappRlyAddr,
		Balance: RollappRlyBalance,
	}, nil
}

func getHubRlyAccData(home string, hd consts.HubData) (*keys.AccountData, error) {
	HubRlyAddr, err := keys.GetRelayerAddress(home, hd.ID)
	if err != nil {
		return nil, err
	}

	HubRlyBalance, err := keys.QueryBalance(
		keys.ChainQueryConfig{
			RPC:    hd.RpcUrl,
			Denom:  consts.Denoms.Hub,
			Binary: consts.Executables.Dymension,
		}, HubRlyAddr,
	)
	if err != nil {
		return nil, err
	}

	return &keys.AccountData{
		Address: HubRlyAddr,
		Balance: HubRlyBalance,
	}, nil
}

func GetHomeDir(home string) string {
	return filepath.Join(home, consts.ConfigDirName.Relayer)
}

func GetConfigFilePath(relayerHome string) string {
	return filepath.Join(relayerHome, "config", "config.yaml")
}
