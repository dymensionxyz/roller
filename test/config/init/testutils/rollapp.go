package testutils

import (
	"path/filepath"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
)

func getRollappKeysDir(root string) string {
	return filepath.Join(root, consts.ConfigDirName.Rollapp, innerKeysDirName)
}

func getHubKeysDir(root string) string {
	return filepath.Join(root, consts.ConfigDirName.HubKeys, innerKeysDirName)
}

func VerifyRollappKeys(root string) error {
	rollappKeysDir := getRollappKeysDir(root)
	sequencerKeyInfoPath := filepath.Join(rollappKeysDir, consts.KeysIds.RollappSequencer+".info")
	if err := verifyFileExists(sequencerKeyInfoPath); err != nil {
		return err
	}
	hubKeysDir := getHubKeysDir(root)
	relayerKeyInfoPath := filepath.Join(hubKeysDir, consts.KeysIds.HubSequencer+".info")
	if err := verifyFileExists(relayerKeyInfoPath); err != nil {
		return err
	}
	err := verifyAndRemoveFilePattern(addressPattern, rollappKeysDir)
	if err != nil {
		return err
	}
	err = verifyAndRemoveFilePattern(addressPattern, hubKeysDir)
	if err != nil {
		return err
	}
	nodeKeyPath := getNodeKeyPath(root)
	if err := verifyFileExists(nodeKeyPath); err != nil {
		return err
	}
	privValKeyPath := getPrivValKeyPath(root)
	if err := verifyFileExists(privValKeyPath); err != nil {
		return err
	}
	return nil
}

func getNodeKeyPath(root string) string {
	return filepath.Join(initconfig.RollappConfigDir(root), "node_key.json")
}

func getPrivValKeyPath(root string) string {
	return filepath.Join(initconfig.RollappConfigDir(root), "priv_validator_key.json")
}

func SanitizeGenesis(genesisPath string) error {
	params := []initconfig.PathValue{
		{
			Path:  "chain_id",
			Value: "PLACEHOLDER_ROLLAPP_ID",
		},
		{
			Path:  "genesis_time",
			Value: "PLACEHOLDER_TIMESTAMP",
		},
		{
			Path:  "app_state.auth.accounts.0.base_account.address",
			Value: "PLACEHOLDER_SEQUENCER_ADDRESS",
		},
		{
			Path:  "app_state.bank.balances.0.address",
			Value: "PLACEHOLDER_SEQUENCER_ADDRESS",
		},
		{
			Path:  "app_state.bank.balances.0.coins.0.amount",
			Value: "PLACEHOLDER_SEQUENCER_BALANCE",
		},
		{
			Path:  "app_state.auth.accounts.1.base_account.address",
			Value: "PLACEHOLDER_RELAYER_ADDRESS",
		},
		{
			Path:  "app_state.bank.balances.1.address",
			Value: "PLACEHOLDER_RELAYER_ADDRESS",
		},
		{
			Path:  "app_state.bank.balances.1.coins.0.amount",
			Value: "PLACEHOLDER_RELAYER_BALANCE",
		},
		{
			Path:  "app_state.sequencers.sequencers.0.operator_address",
			Value: "PLACEHOLDER_OPERATOR_ADDRESS",
		},
		{
			Path:  "app_state.sequencers.sequencers.0.consensus_pubkey.key",
			Value: "PLACEHOLDER_CONSENSUS_PUBKEY",
		},
		{
			Path:  "app_state.genutil.gen_txs.0.body.messages.0.delegator_address",
			Value: "PLACEHOLDER_SEQUENCER_ADDRESS",
		},
		{
			Path:  "app_state.genutil.gen_txs.0.body.messages.0.pubkey.key",
			Value: "PLACEHOLDER_SEQUENCER_PUBKEY",
		},
		{
			Path:  "app_state.genutil.gen_txs.0.body.messages.0.validator_address",
			Value: "PLACEHOLDER_SEQUENCER_GOVERNER_ADDRESS",
		},
		{
			Path:  "app_state.genutil.gen_txs.0.signatures.0",
			Value: "PLACEHOLDER_SEQUENCER_SIGNATURE",
		},
		{
			Path:  "app_state.genutil.gen_txs.0.body.memo",
			Value: "PLACEHOLDER_GENTX_MEMO",
		},
		{
			Path:  "app_state.genutil.gen_txs.0.auth_info.signer_infos.0.public_key.key",
			Value: "PLACEHOLDER_SEQUENCER_PUBKEY",
		},
	}

	err := initconfig.UpdateJSONParams(genesisPath, params)
	if err != nil {
		return err
	}
	return nil
}
