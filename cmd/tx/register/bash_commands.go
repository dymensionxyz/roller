package register

import (
	"os/exec"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"

	"fmt"

	"github.com/dymensionxyz/roller/cmd/consts"
)

// TODO: create tokenmetadata.json
func getRegisterRollappCmd(rollappConfig config.RollappConfig) *exec.Cmd {
	tokenMetadataPath := filepath.Join(rollappConfig.Home, consts.ConfigDirName.Rollapp, "config", "tokenmetadata.json")
	cmdArgs := []string{
		"tx", "rollapp", "create-rollapp", rollappConfig.RollappID, "3", `{"Addresses":[]}`, tokenMetadataPath,
	}
	cmdArgs = append(cmdArgs, getCommonDymdTxFlags(rollappConfig)...)
	return exec.Command(
		consts.Executables.Dymension, cmdArgs...,
	)
}

func getRegisterSequencerCmd(rollappConfig config.RollappConfig) (*exec.Cmd, error) {
	seqPubKey, err := utils.GetSequencerPubKey(rollappConfig)
	if err != nil {
		return nil, err
	}
	description := fmt.Sprintf(`{"Moniker":"%s","Identity":"","Website":"","SecurityContact":"","Details":""}`,
		consts.KeysIds.HubSequencer)
	cmdArgs := []string{
		"tx", "sequencer", "create-sequencer",
		seqPubKey,
		rollappConfig.RollappID,
		description,
	}
	cmdArgs = append(cmdArgs, getCommonDymdTxFlags(rollappConfig)...)
	return exec.Command(consts.Executables.Dymension, cmdArgs...), nil
}

func getCommonDymdTxFlags(rollappConfig config.RollappConfig) []string {
	commonFlags := utils.GetCommonDymdFlags(rollappConfig)
	txArgs := []string{
		"--from", consts.KeysIds.HubSequencer,
		"--keyring-backend", "test",
		"--keyring-dir", filepath.Join(rollappConfig.Home, consts.ConfigDirName.HubKeys),
		"--yes", "--broadcast-mode", "block", "--chain-id", rollappConfig.HubData.ID,
		"--gas-prices", rollappConfig.HubData.GAS_PRICE + consts.Denoms.Hub,
	}
	return append(commonFlags, txArgs...)
}
