package testutils

import (
	"github.com/dymensionxyz/roller/config"
	"github.com/dymensionxyz/roller/relayer"
	"github.com/dymensionxyz/roller/utils"
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
	return relayer.WriteRlyConfig(rlpCfg.Home, rlyCfg)
}
