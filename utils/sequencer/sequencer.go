package sequencer

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	cosmossdkmath "cosmossdk.io/math"
	cosmossdktypes "github.com/cosmos/cosmos-sdk/types"
	dymrollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/tx/tx_utils"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/rollapp"
	"github.com/dymensionxyz/roller/utils/roller"
	"github.com/dymensionxyz/roller/utils/tx"
)

func GetSequencerConfigDir(rollerHome string) string {
	return filepath.Join(rollerHome, consts.ConfigDirName.Rollapp, "config")
}

func Register(raCfg roller.RollappConfig, desiredBond cosmossdktypes.Coin) error {
	seqPubKey, err := keys.GetSequencerPubKey(raCfg)
	if err != nil {
		return err
	}

	seqMetadataPath := filepath.Join(
		raCfg.Home,
		consts.ConfigDirName.Rollapp,
		"init",
		"sequencer-metadata.json",
	)
	_, err = isValidSequencerMetadata(seqMetadataPath)
	if err != nil {
		return err
	}

	home := roller.GetRootDir()

	customRewardAddress, _ := pterm.DefaultInteractiveTextInput.WithDefaultText(
		"would you like to use a custom reward address (leave empty to use the sequencer address)",
	).Show()
	customRewardAddress = strings.TrimSpace(customRewardAddress)

	args := []string{
		"tx",
		"sequencer",
		"create-sequencer",
		seqPubKey,
		raCfg.RollappID,
		desiredBond.String(),
		seqMetadataPath,
		"--from", consts.KeysIds.HubSequencer,
		"--keyring-backend", string(raCfg.KeyringBackend),
		"--fees", fmt.Sprintf("%d%s", consts.DefaultTxFee, consts.Denoms.Hub),
		"--gas", "auto",
		"--gas-adjustment", "1.3",
		"--keyring-dir", filepath.Join(home, consts.ConfigDirName.HubKeys),
		"--node", raCfg.HubData.RpcUrl, "--chain-id", raCfg.HubData.ID,
	}
	if customRewardAddress != "" {
		cArgs := []string{"--reward-address", customRewardAddress}
		args = append(args, cArgs...)
	}

	displayBond, err := BaseDenomToDenom(desiredBond, 18)
	if err != nil {
		return err
	}

	pswFileName, err := filesystem.GetOsKeyringPswFileName(consts.Executables.Dymension)
	if err != nil {
		return err
	}
	fp := filepath.Join(home, string(pswFileName))
	psw, err := filesystem.ReadFromFile(fp)
	if err != nil {
		return err
	}

	automaticPrompts := map[string]string{
		"Enter keyring passphrase":    psw,
		"Re-enter keyring passphrase": psw,
	}
	manualPromptResponses := map[string]string{
		"signatures": fmt.Sprintf(
			"this transaction is going to register your sequencer with %s bond. do you want to continue?",
			pterm.Yellow(pterm.Bold.Sprint(displayBond.String())),
		),
	}

	txOutput, err := bash.ExecuteCommandWithPromptHandler(
		consts.Executables.Dymension,
		args,
		automaticPrompts,
		manualPromptResponses,
	)
	if err != nil {
		return err
	}

	err = tx_utils.CheckTxYamlStdOut(*txOutput)
	if err != nil {
		return err
	}

	txHash, err := bash.ExtractTxHash(txOutput.String())
	if err != nil {
		return err
	}

	err = tx.MonitorTransaction(raCfg.HubData.RpcUrl, txHash)
	if err != nil {
		return err
	}

	return nil
}

func CanSequencerBeRegisteredForRollapp(raID string, hd consts.HubData) (bool, error) {
	raResponse, err := rollapp.GetMetadataFromChain(raID, hd)
	if err != nil {
		return false, err
	}

	if raResponse.Rollapp.InitialSequencer == "" {
		return false, nil
	}

	return true, nil
}

func isValidSequencerMetadata(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}

	// nolint:errcheck
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return false, err
	}

	var sm dymrollapptypes.Rollapp
	err = json.Unmarshal(b, &sm)
	if err != nil {
		return false, err
	}

	return true, err
}

func GetSequencerAccountAddress(cfg roller.RollappConfig) (string, error) {
	kc := keys.KeyConfig{
		ChainBinary:    consts.Executables.Dymension,
		ID:             consts.KeysIds.HubSequencer,
		Dir:            consts.ConfigDirName.HubKeys,
		KeyringBackend: cfg.KeyringBackend,
	}

	seqAddr, err := kc.Address(cfg.Home)
	if err != nil {
		return "", err
	}

	return seqAddr, nil
}

func GetRpcEndpointFromChain(raID string, hd consts.HubData) (string, error) {
	seqAddr, err := rollapp.GetCurrentProposer(raID, hd)
	if err != nil {
		return "", err
	}

	if seqAddr == "" {
		return "", fmt.Errorf("no proposer found for rollapp %s", raID)
	}

	metadata, err := GetMetadata(seqAddr, hd)
	if err != nil {
		return "", err
	}

	return metadata.Rpcs[0], err
}

func GetRestEndpointFromChain(raID string, hd consts.HubData) (string, error) {
	seqAddr, err := rollapp.GetCurrentProposer(raID, hd)
	if err != nil {
		return "", err
	}

	if seqAddr == "" {
		return "", fmt.Errorf("no proposer found for rollapp %s", raID)
	}

	metadata, err := GetMetadata(seqAddr, hd)
	if err != nil {
		return "", err
	}

	return metadata.RestApiUrls[0], err
}

func GetJsonRpcEndpointFromChain(raID string, hd consts.HubData) (string, error) {
	seqAddr, err := rollapp.GetCurrentProposer(raID, hd)
	if err != nil {
		return "", err
	}

	if seqAddr == "" {
		return "", fmt.Errorf("no proposer found for rollapp %s", raID)
	}

	metadata, err := GetMetadata(seqAddr, hd)
	if err != nil {
		return "", err
	}

	return metadata.RestApiUrls[0], err
}

func GetMinSequencerBondInBaseDenom(raID string, hd consts.HubData) (*cosmossdktypes.Coin, error) {
	var qra struct {
		Rollapp dymrollapptypes.Rollapp `json:"rollapp"`
	}
	cmd := exec.Command(
		consts.Executables.Dymension,
		"q", "rollapp", "show", raID, "-o", "json", "--node", hd.RpcUrl, "--chain-id", hd.ID,
	)

	out, err := bash.ExecCommandWithStdout(cmd)
	if err != nil {
		return nil, err
	}

	_ = json.Unmarshal(out.Bytes(), &qra)

	var c cosmossdktypes.Coin
	if qra.Rollapp.MinSequencerBond != nil {
		c = qra.Rollapp.MinSequencerBond[0]
	} else {
		params, err := rollapp.GetRollappParams(hd)
		if err != nil {
			return nil, err
		}
		c = params.Params.MinSequencerBondGlobal
	}

	return &c, nil
}

func BaseDenomToDenom(
	coin cosmossdktypes.Coin,
	exponent int,
) (cosmossdktypes.Coin, error) {
	exp := cosmossdkmath.NewIntWithDecimal(1, exponent)

	coin.Amount = coin.Amount.Quo(exp)
	coin.Denom = coin.Denom[1:]

	return coin, nil
}

func DenomToBaseDenom(
	coin cosmossdktypes.Coin,
	exponent int,
) (cosmossdktypes.Coin, error) {
	exp := cosmossdkmath.NewIntWithDecimal(1, exponent)

	coin.Amount = coin.Amount.Mul(exp)

	return coin, nil
}

// TODO: dymd q sequencer show-sequencer could be used instead
func IsRegisteredAsSequencer(seq []Info, addr string) bool {
	if len(seq) == 0 {
		return false
	}

	return slices.ContainsFunc(
		seq,
		func(s Info) bool { return strings.Compare(s.Address, addr) == 0 },
	)
}

func GetLatestSnapshot(raID string, hd consts.HubData) (*SnapshotInfo, error) {
	sequencers, err := RegisteredRollappSequencersOnHub(raID, hd)
	if err != nil {
		return nil, err
	}

	var latestSnapshot *SnapshotInfo
	maxHeight := 0

	for _, s := range sequencers.Sequencers {
		for _, snapshot := range s.Metadata.Snapshots {
			height, err := strconv.Atoi(snapshot.Height)
			if err != nil {
				continue
			}

			if height > maxHeight {
				maxHeight = height
				latestSnapshot = snapshot
			}
		}
	}

	return latestSnapshot, nil
}

func GetAllP2pPeers(raID string, hd consts.HubData) ([]string, error) {
	sequencers, err := RegisteredRollappSequencersOnHub(raID, hd)
	if err != nil {
		return nil, err
	}

	var peers []string

	for _, s := range sequencers.Sequencers {
		if s.OptedIn && s.Status == "OPERATING_STATUS_BONDED" {
			peers = append(peers, s.Metadata.P2PSeeds...)
		}
	}

	return peers, nil
}

func RegisteredRollappSequencersOnHub(
	raID string, hd consts.HubData,
) (*Sequencers, error) {
	var seq Sequencers
	cmd := getShowSequencerByRollappCmd(raID, hd)

	out, err := bash.ExecCommandWithStdout(cmd)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(out.Bytes(), &seq)
	if err != nil {
		return nil, err
	}

	return &seq, nil
}

func RegisteredRollappSequencers(
	raID string,
) (*Sequencers, error) {
	var seq Sequencers
	cmd := getShowSequencersCmd(raID)

	out, err := bash.ExecCommandWithStdout(cmd)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(out.Bytes(), &seq)
	if err != nil {
		return nil, err
	}

	return &seq, nil
}

func GetMetadata(
	addr string,
	hd consts.HubData,
) (*Metadata, error) {
	var seqinfo ShowSequencerResponse

	cmd := exec.Command(
		consts.Executables.Dymension,
		"q", "sequencer", "show-sequencer", addr,
		"--node", hd.RpcUrl, "-o", "json", "--chain-id", hd.ID,
	)

	out, err := bash.ExecCommandWithStdout(cmd)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(out.Bytes(), &seqinfo)
	if err != nil {
		return nil, err
	}

	return &seqinfo.Sequencer.Metadata, nil
}

func getShowSequencerByRollappCmd(raID string, hd consts.HubData) *exec.Cmd {
	return exec.Command(
		consts.Executables.Dymension,
		"q", "sequencer", "show-sequencers-by-rollapp",
		raID, "-o", "json", "--node", hd.RpcUrl, "--chain-id", hd.ID,
	)
}

func getShowSequencersCmd(raID string) *exec.Cmd {
	return exec.Command(
		consts.Executables.RollappEVM,
		"q", "sequencers", "sequencers",
		"-o", "json", "--node", consts.DefaultRollappRPC, "--chain-id", raID,
	)
}

func GetHubSequencerAddress(cfg roller.RollappConfig) (string, error) {
	kc := keys.KeyConfig{
		ChainBinary:    consts.Executables.Dymension,
		ID:             consts.KeysIds.HubSequencer,
		Dir:            consts.ConfigDirName.HubKeys,
		KeyringBackend: cfg.KeyringBackend,
	}
	seqAddr, err := kc.Address(cfg.Home)
	if err != nil {
		return "", err
	}

	return seqAddr, nil
}

func GetSequencerData(cfg roller.RollappConfig) ([]keys.AccountData, error) {
	seqAddr, err := GetHubSequencerAddress(cfg)
	if err != nil {
		return nil, err
	}

	sequencerBalance, err := keys.QueryBalance(
		keys.ChainQueryConfig{
			Binary: consts.Executables.Dymension,
			Denom:  consts.Denoms.Hub,
			RPC:    cfg.HubData.RpcUrl,
		}, seqAddr,
	)
	if err != nil {
		return nil, err
	}
	return []keys.AccountData{
		{
			Address: seqAddr,
			Balance: *sequencerBalance,
		},
	}, nil
}

func GetSequencerBond(address string, hd consts.HubData) (*cosmossdktypes.Coins, error) {
	c := exec.Command(
		consts.Executables.Dymension,
		"q",
		"sequencer",
		"show-sequencer",
		address,
		"--output",
		"json",
		"--node", hd.RpcUrl,
		"--chain-id", hd.ID,
	)

	var GetSequencerResponse ShowSequencerResponse
	out, err := bash.ExecCommandWithStdout(c)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(out.Bytes(), &GetSequencerResponse)
	if err != nil {
		return nil, err
	}

	return &GetSequencerResponse.Sequencer.Tokens, nil
}

func GetDymintFilePath(root string) string {
	return filepath.Join(root, consts.ConfigDirName.Rollapp, "config", "dymint.toml")
}

func GetAppConfigFilePath(root string) string {
	return filepath.Join(root, consts.ConfigDirName.Rollapp, "config", "app.toml")
}

func GetWhitelistedRelayersOnHub(address string, hd consts.HubData) ([]string, error) {
	seqResp, err := showSequencer(address, hd)
	if err != nil {
		return nil, err
	}

	return seqResp.Sequencer.WhitelistedRelayers, nil
}

func getShowSequencerCmd(addr string, hd consts.HubData) *exec.Cmd {
	return exec.Command(
		consts.Executables.Dymension,
		"q",
		"sequencer",
		"show-sequencer",
		addr,
		"--output",
		"json",
		"--node", hd.RpcUrl,
		"--chain-id", hd.ID,
	)
}

func showSequencer(addr string, hd consts.HubData) (*ShowSequencerResponse, error) {
	c := getShowSequencerCmd(addr, hd)
	var GetSequencerResponse ShowSequencerResponse
	out, err := bash.ExecCommandWithStdout(c)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(out.Bytes(), &GetSequencerResponse)
	if err != nil {
		return nil, err
	}
	return &GetSequencerResponse, nil
}

func CheckExistingSequencer(home string) (*CheckExistingSequencerResponse, error) {
	rollerData, err := roller.LoadConfig(home)
	if err != nil {
		return nil, err
	}

	pterm.Info.Printfln("checking for existing sequencer keys")
	kcs := keys.GetSequencerKeysConfig(rollerData.KeyringBackend)
	kc := kcs[0]

	ok, err := kc.IsInKeyring(home)
	if err != nil {
		return nil, err
	}
	if ok {
		pterm.Info.Printfln("sequencer key already exists, verifying...")
		ki, err := kc.Info(home)
		if err != nil {
			return nil, err
		}

		existingSeqs, err := RegisteredRollappSequencersOnHub(
			rollerData.RollappID,
			rollerData.HubData,
		)
		if err != nil {
			return nil, err
		}

		isSequencer := IsRegisteredAsSequencer(existingSeqs.Sequencers, ki.Address)
		if isSequencer {
			cp, err := rollapp.GetCurrentProposer(rollerData.RollappID, rollerData.HubData)
			if err != nil {
				return nil, err
			}

			if cp == ki.Address {
				pterm.Warning.Printfln(
					"sequencer key already exists and is registered on the hub as proposer",
				)
				return &CheckExistingSequencerResponse{
					IsSequencerAlreadyRegistered: true,
					IsSequencerKeyPresent:        true,
					IsSequencerProposer:          true,
				}, nil
			}
			pterm.Warning.Printfln("sequencer key already exists and is registered on the hub")
			return &CheckExistingSequencerResponse{
				IsSequencerAlreadyRegistered: true,
				IsSequencerKeyPresent:        true,
				IsSequencerProposer:          false,
			}, nil
		} else {
			pterm.Info.Printfln("sequencer key already exists but is not registered on the hub")
		}
	}

	return &CheckExistingSequencerResponse{
		IsSequencerAlreadyRegistered: false,
		IsSequencerKeyPresent:        false,
		IsSequencerProposer:          false,
	}, nil
}

func UpdateWhitelistedRelayers(
	home, raRelayerAddress, kb string,
	hd consts.HubData,
) error {
	args := []string{
		"tx",
		"sequencer",
		"update-whitelisted-relayers",
		raRelayerAddress,
		"--from", consts.KeysIds.HubSequencer,
		"--home", filepath.Join(home, consts.ConfigDirName.HubKeys),
		"--keyring-backend", kb,
		"--chain-id", hd.ID,
		"--node", hd.RpcUrl,
		"--fees", fmt.Sprintf("%d%s", consts.DefaultTxFee, consts.Denoms.Hub),
	}

	psw, err := filesystem.ReadOsKeyringPswFile(home, consts.Executables.Dymension)
	if err != nil {
		return err
	}

	automaticPrompts := map[string]string{
		"Enter keyring passphrase": psw,
	}
	manualPromptResponses := map[string]string{
		"signatures": "this transaction is going to update the whitelisted relayers. do you want to continue?",
	}

	txOutput, err := bash.ExecuteCommandWithPromptHandler(
		consts.Executables.Dymension,
		args,
		automaticPrompts,
		manualPromptResponses,
	)
	if err != nil {
		return err
	}

	txHash, err := bash.ExtractTxHash(txOutput.String())
	if err != nil {
		return err
	}

	err = tx.MonitorTransaction(hd.RpcUrl, txHash)
	if err != nil {
		return err
	}

	return nil
}

func GetSequencerOperatorAddress(home string, kb string) (string, error) {
	rollappConfigDirPath := filepath.Join(home, consts.ConfigDirName.HubKeys)
	args := []string{
		"keys",
		"show",
		consts.KeysIds.HubSequencer,
		"-a",
		"--keyring-backend",
		kb,
		"--home",
		rollappConfigDirPath,
		"--bech",
		"val",
	}
	psw, err := filesystem.ReadOsKeyringPswFile(home, consts.Executables.Dymension)
	if err != nil {
		return "", err
	}

	automaticPrompts := map[string]string{
		"Enter keyring passphrase": psw,
	}

	addr, err := bash.ExecuteCommandWithPromptHandler(
		consts.Executables.RollappEVM,
		args,
		automaticPrompts,
		nil,
	)
	if err != nil {
		fmt.Println("val addr failed")
		return "", err
	}

	a := strings.TrimSpace(addr.String())
	return a, nil
}

type RaWhitelisterRelayersResponse struct {
	Relayers []string `json:"relayers"`
}

func GetWhitelistedRelayersOnRa(raOpAddr string) ([]string, error) {
	cmd := exec.Command(
		consts.Executables.RollappEVM,
		"q",
		"sequencers",
		"whitelisted-relayers",
		raOpAddr,
		"--output",
		"json",
	)

	o, err := bash.ExecCommandWithStdout(cmd)
	if err != nil {
		return nil, err
	}

	var response RaWhitelisterRelayersResponse
	if err := json.Unmarshal(o.Bytes(), &response); err != nil {
		return nil, err
	}

	return response.Relayers, nil
}
