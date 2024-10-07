package set

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/roller"
)

func setDA(rlpCfg roller.RollappConfig, value string) error {
	daValue := consts.DAType(value)
	if daValue == rlpCfg.DA.Backend {
		return nil
	}

	if !roller.IsValidDAType(value) {
		return fmt.Errorf("invalid DA type. Supported types are: %v", roller.SupportedDas)
	}
	return updateDaConfig(rlpCfg, daValue)
}

func updateDaConfig(rlpCfg roller.RollappConfig, newDa consts.DAType) error {
	daCfgDirPath := filepath.Join(rlpCfg.Home, consts.ConfigDirName.DALightNode)
	dirExist, err := filesystem.DirNotEmpty(daCfgDirPath)
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

	rlpCfg.DA.Backend = newDa
	if err := sequencer.UpdateDymintDAConfig(rlpCfg); err != nil {
		return err
	}

	if err := roller.WriteConfig(rlpCfg); err != nil {
		return err
	}

	fmt.Printf("💈 RollApp DA has been successfully set to '%s'\n\n", newDa)
	if newDa != consts.Local {
		addresses := make([]keys.KeyInfo, 0)
		damanager := datalayer.NewDAManager(newDa, rlpCfg.Home)
		daAddress, err := damanager.GetDAAccountAddress()
		if err != nil {
			return err
		}
		addresses = append(
			addresses, keys.KeyInfo{
				Name:    damanager.GetKeyName(),
				Address: daAddress.Address,
			},
		)

		keys.PrintAddressesWithTitle(addresses)
		fmt.Printf("\n🔔 Please fund this address to run the DA light client.\n")
	}
	return nil
}
