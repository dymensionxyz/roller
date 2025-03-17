package gov

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	paramsutils "github.com/cosmos/cosmos-sdk/x/params/client/utils"
	"github.com/dymensionxyz/roller/cmd/consts"
	bashutils "github.com/dymensionxyz/roller/utils/bash"
	"github.com/pterm/pterm"
)

type CosmosTx struct {
	TxHash string `json:"txhash"`
	Code   int    `json:"code"`
	RawLog string `json:"raw_log"`
}

// ParamChangeProposal submits a param change proposal to the chain, signed by keyName.
func ParamChangeProposal(home, keyName, keyring string, prop *paramsutils.ParamChangeProposalJSON) (string, error) {
	content, err := json.Marshal(prop)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(content)
	proposalFilename := fmt.Sprintf("%x.json", hash)
	proposalPath := filepath.Join(home, proposalFilename)

	err = os.WriteFile(proposalPath, content, 0o644)
	if err != nil {
		pterm.Error.Printf("Error writing prop file: %v\n", err)
		return "", err
	}

	cmd := exec.Command(
		consts.Executables.RollappEVM,
		"tx", "gov", "submit-legacy-proposal",
		"param-change",
		proposalPath,
		"--gas", "auto",
		"--from", keyName,
		"--keyring-backend", keyring,
		"--output", "json",
		"-y",
	)

	out, err := bashutils.ExecCommandWithStdout(cmd)
	if err != nil {
		return "", err
	}

	output := CosmosTx{}
	err = json.Unmarshal(out.Bytes(), &output)
	if err != nil {
		return "", err
	}
	if output.Code != 0 {
		return output.TxHash, fmt.Errorf("transaction failed with code %d: %s", output.Code, output.RawLog)
	}

	return output.TxHash, nil
}
