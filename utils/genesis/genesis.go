package genesis

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cometbft/cometbft/types"
	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	globalutils "github.com/dymensionxyz/roller/utils"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config"
	"github.com/dymensionxyz/roller/utils/rollapp"
	"github.com/dymensionxyz/roller/utils/sequencer"
)

type AppState struct {
	Bank Bank `json:"bank"`
}

type Bank struct {
	Supply []Denom `json:"supply"`
}

type Denom struct {
	Denom string `json:"denom"`
}

func DownloadGenesis(home string, rollappConfig config.RollappConfig) error {
	pterm.Info.Println("downloading genesis file")

	genesisPath := GetGenesisFilePath(home)
	genesisUrl := rollappConfig.GenesisUrl
	if genesisUrl == "" {
		return fmt.Errorf("RollApp's genesis url field is empty, contact the rollapp owner")
	}

	err := globalutils.DownloadFile(genesisUrl, genesisPath)
	if err != nil {
		return err
	}

	// move to helper function with a spinner?
	genesis, err := types.GenesisDocFromFile(genesisPath)
	if err != nil {
		return err
	}

	if genesis.ChainID != rollappConfig.RollappID {
		err = fmt.Errorf(
			"the genesis file ChainID (%s) does not match  the rollapp ID you're trying to initialize ("+
				"%s)",
			genesis.ChainID,
			rollappConfig.RollappID,
		)
		return err
	}

	return nil
}

func calculateSHA256(filepath string) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", fmt.Errorf("error opening file: %v", err)
	}
	// nolint:errcheck
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("error calculating hash: %v", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func getRollappGenesisHash(raID string, hd consts.HubData) (string, error) {
	var raResponse rollapp.ShowRollappResponse
	getRollappCmd := exec.Command(
		consts.Executables.Dymension,
		"q", "rollapp", "show",
		raID, "-o", "json", "--node", hd.RPC_URL, "--chain-id", hd.ID,
	)

	out, err := bash.ExecCommandWithStdout(getRollappCmd)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(out.Bytes(), &raResponse)
	if err != nil {
		return "", err
	}
	return raResponse.Rollapp.GenesisInfo.GenesisChecksum, nil
}

func CompareGenesisChecksum(root, raID string, hd consts.HubData) (bool, error) {
	genesisPath := GetGenesisFilePath(root)
	downloadedGenesisHash, err := calculateSHA256(genesisPath)
	if err != nil {
		pterm.Error.Println("failed to calculate hash of genesis file: ", err)
		return false, err
	}
	raGenesisHash, _ := getRollappGenesisHash(raID, hd)
	if downloadedGenesisHash != raGenesisHash {
		err = fmt.Errorf(
			"the hash of the downloaded file (%s) does not match the one registered with the rollapp (%s)",
			downloadedGenesisHash,
			raGenesisHash,
		)
		return false, err
	}

	return true, nil
}

func CompareRollappArchiveChecksum(
	filepath string,
	si sequencer.SnapshotInfo,
) (bool, error) {
	downloadedGenesisHash, err := calculateSHA256(filepath)
	if err != nil {
		pterm.Error.Println("failed to calculate hash of genesis file: ", err)
		return false, err
	}
	onChainHash := si.Checksum
	if downloadedGenesisHash != onChainHash {
		err = fmt.Errorf(
			"the hash of the downloaded file (%s) does not match the one registered with the rollapp (%s)",
			downloadedGenesisHash,
			onChainHash,
		)
		return false, err
	}

	return true, nil
}

func GetGenesisFilePath(root string) string {
	return filepath.Join(
		rollapp.RollappConfigDir(root),
		"genesis.json",
	)
}
