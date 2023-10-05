package testutils

import (
	"errors"
	"github.com/dymensionxyz/roller/config"
	"github.com/dymensionxyz/roller/relayer"
	"github.com/dymensionxyz/roller/utils"
	"reflect"
)

func SanitizeRlyConfig(rlpCfg *config.RollappConfig) error {
	rlyCfg, err := relayer.ReadRlyConfig(rlpCfg.Home)
	if err != nil {
		return err
	}
	const placeholderRollappID = "PLACEHOLDER_ROLLAPP_ID"
	err = utils.SetNestedValue(rlyCfg, []string{"chains", rlpCfg.RollappID, "value", "chain-id"}, placeholderRollappID)
	if err != nil {
		return err
	}
	rlpData, err := utils.GetNestedValue(rlyCfg, []string{"chains", rlpCfg.RollappID})
	if err != nil {
		return err
	}
	err = utils.SetNestedValue(rlyCfg, []string{"chains", placeholderRollappID}, rlpData)
	if err != nil {
		return err
	}
	err = utils.SetNestedValue(rlyCfg, []string{"chains", rlpCfg.RollappID}, nil)
	if err != nil {
		return err
	}
	err = utils.SetNestedValue(rlyCfg, []string{"paths", "rollapp-hub", "dst", "chain-id"}, placeholderRollappID)
	if err != nil {
		return err
	}
	return relayer.WriteRlyConfig(rlpCfg.Home, rlyCfg)
}

func VerifyRlyConfig(rollappConfig config.RollappConfig, goldenDirPath string) error {
	goldenRlyCfg, err := relayer.ReadRlyConfig(goldenDirPath)
	if err != nil {
		return err
	}
	rlyCfg, err := relayer.ReadRlyConfig(rollappConfig.Home)
	if err != nil {
		return err
	}
	if reflect.DeepEqual(rlyCfg, goldenRlyCfg) {
		return nil
	}
	return errors.New("rly config does not match golden config")
}
