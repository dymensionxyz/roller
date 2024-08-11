package set

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	"github.com/dymensionxyz/roller/sequencer"
	globalutils "github.com/dymensionxyz/roller/utils"
	config2 "github.com/dymensionxyz/roller/utils/config"
	"github.com/dymensionxyz/roller/utils/config/toml"
)

func setDA(rlpCfg config2.RollappConfig, value string) error {
	daValue := consts.DAType(value)
	if daValue == rlpCfg.DA {
		return nil
	}

	if !config2.IsValidDAType(value) {
		return fmt.Errorf("invalid DA type. Supported types are: %v", config2.SupportedDas)
	}
	return updateDaConfig(rlpCfg, daValue)
}

func updateDaConfig(rlpCfg config2.RollappConfig, newDa consts.DAType) error {
	daCfgDirPath := filepath.Join(rlpCfg.Home, consts.ConfigDirName.DALightNode)
	dirExist, err := globalutils.DirNotEmpty(daCfgDirPath)
	if err != nil {
		return err
	}

	if dirExist {
		if yes, err := utils.PromptBool("Changing DA will remove the old DA keys permanently. Are you sure you want to proceed"); err != nil {
			return err
		} else if !yes {
			return nil
		}
	}
	if err := os.RemoveAll(daCfgDirPath); err != nil {
		return err
	}

	daManager := datalayer.NewDAManager(newDa, rlpCfg.Home)
	_, err = daManager.InitializeLightNodeConfig()
	if err != nil {
		return err
	}

	rlpCfg.DA = newDa
	if err := sequencer.UpdateDymintDAConfig(rlpCfg); err != nil {
		return err
	}

	if err := toml.WriteConfigToTOML(rlpCfg); err != nil {
		return err
	}

	fmt.Printf("💈 RollApp DA has been successfully set to '%s'\n\n", newDa)
	if newDa != consts.Local {
		addresses := make([]utils.KeyInfo, 0)
		damanager := datalayer.NewDAManager(newDa, rlpCfg.Home)
		daAddress, err := damanager.GetDAAccountAddress()
		if err != nil {
			return err
		}
		addresses = append(
			addresses, utils.KeyInfo{
				Name:    damanager.GetKeyName(),
				Address: daAddress.Address,
			},
		)

		utils.PrintAddressesWithTitle(addresses)
		fmt.Printf("\n🔔 Please fund this address to run the DA light client.\n")
	}
	return nil
}
