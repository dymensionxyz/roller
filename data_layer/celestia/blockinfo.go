package celestia

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config"
)

// GetLatestDABlock returns the latest DA (Data Availability) block information.
// It executes the CelestiaApp command "q block --node" to retrieve the block data.
// It then extracts the block height and block ID hash from the JSON response.
// Returns the block height, block ID hash, and any error encountered during the process.
func GetLatestBlock(raCfg config.RollappConfig) (string, string, error) {
	cmd := exec.Command(
		consts.Executables.CelestiaApp,
		"q", "block", "--node", raCfg.DA.RpcUrl, "--chain-id", string(raCfg.DA.ID),
	)

	out, err := bash.ExecCommandWithStdout(cmd)
	if err != nil {
		return "", "", err
	}

	var tx map[string]interface{}
	err = json.Unmarshal(out.Bytes(), &tx)
	if err != nil {
		return "", "", err
	}

	// Access tx.Block.Header.Height
	var height string
	if block, ok := tx["block"].(map[string]interface{}); ok {
		if header, ok := block["header"].(map[string]interface{}); ok {
			if h, ok := header["height"].(string); ok {
				height = h
			}
		}
	}

	// Access tx.BlockId.Hash
	var blockIdHash string
	if blockId, ok := tx["block_id"].(map[string]interface{}); ok {
		if hash, ok := blockId["hash"].(string); ok {
			blockIdHash = hash
		}
	}
	err = json.Unmarshal(out.Bytes(), &tx)
	if err != nil {
		return "", "", err
	}

	return height, blockIdHash, nil
}

// GetDABlockByHeight returns the DA (Data Availability) block information for the given height.
// It executes the CelestiaApp command "q block <height> --node" to retrieve the block data,
// where <height> is the input parameter.
// It then extracts the block height and block ID hash from the JSON response.
// Returns the block height, block ID hash, and any error encountered during the process.
func GetBlockByHeight(h string, raCfg config.RollappConfig) (string, string, error) {
	cmd := exec.Command(
		consts.Executables.CelestiaApp,
		"q", "block", h, "--node", raCfg.DA.RpcUrl, "--chain-id", string(raCfg.DA.ID),
	)

	out, err := bash.ExecCommandWithStdout(cmd)
	if err != nil {
		return "", "", err
	}

	var tx map[string]interface{}
	err = json.Unmarshal(out.Bytes(), &tx)
	if err != nil {
		return "", "", err
	}

	// Access tx.Block.Header.Height
	var height string
	if block, ok := tx["block"].(map[string]interface{}); ok {
		if header, ok := block["header"].(map[string]interface{}); ok {
			if h, ok := header["height"].(string); ok {
				height = h
			}
		}
	}

	// Access tx.BlockId.Hash
	var blockIdHash string
	if blockId, ok := tx["block_id"].(map[string]interface{}); ok {
		if hash, ok := blockId["hash"].(string); ok {
			blockIdHash = hash
		}
	}
	err = json.Unmarshal(out.Bytes(), &tx)
	if err != nil {
		return "", "", err
	}

	return height, blockIdHash, nil
}

// ExtractHeightfromDAPath function extracts the celestia height from DA path that's
// available on the hub
func ExtractHeightfromDAPath(input string) (string, error) {
	parts := strings.Split(input, "|")
	if len(parts) < 2 {
		return "", fmt.Errorf("input string does not have enough parts")
	}
	return parts[1], nil
}
