package testutils

import (
	"errors"
	"fmt"
	"github.com/dymensionxyz/roller/config"
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/utils"
	"github.com/pelletier/go-toml"
	"os"

	"github.com/dymensionxyz/roller/cmd/consts"
	"path/filepath"
	"regexp"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
)

const innerKeysDirName = "keyring-test"
const addressPattern = `.*\.address`

func SanitizeConfigDir(root string, rlpCfg *config.RollappConfig) error {
	dirsToClean := []string{getLightNodeKeysDir(root), getRelayerKeysDir(root), getRollappKeysDir(root),
		getHubKeysDir(root), filepath.Join(root, consts.ConfigDirName.LocalHub)}
	for _, dir := range dirsToClean {
		if err := os.RemoveAll(dir); err != nil {
			return err
		}
	}
	privValKeyPath := getPrivValKeyPath(root)
	if err := os.Remove(privValKeyPath); err != nil {
		return err
	}
	nodeKeyPath := getNodeKeyPath(root)
	if err := os.Remove(nodeKeyPath); err != nil {
		return err
	}
	if err := SanitizeGenesis(initconfig.GetGenesisFilePath(root)); err != nil {
		return err
	}
	if err := SanitizeRlyConfig(rlpCfg); err != nil {
		return err
	}

	if err := SanitizeDymintToml(root); err != nil {
		return err
	}
	return nil
}

func SanitizeDymintToml(root string) error {
	dymintTomlPath := sequencer.GetDymintFilePath(root)
	var tomlCfg, err = toml.LoadFile(dymintTomlPath)
	if err != nil {
		return fmt.Errorf("failed to load %s: %v", dymintTomlPath, err)
	}
	tomlCfg.Set("keyring_home_dir", "PLACEHOLDER_KEYRING_HOME_DIR")
	tomlCfg.Set("namespace_id", "PLACEHOLDER_NAMESPACE_ID")
	tomlCfg.Set("da_config", "PLACEHOLDER_DA_CONFIG")
	return utils.WriteTomlTreeToFile(tomlCfg, dymintTomlPath)
}
func verifyFileExists(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		return errors.New("File " + path + " does not exist")
	}
	return nil
}

func FileWithPatternPath(dir, pattern string) (string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	r, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}

	for _, file := range files {
		if r.MatchString(file.Name()) {
			return filepath.Join(dir, file.Name()), nil
		}
	}

	return "", nil
}

func verifyAndRemoveFilePattern(pattern string, dir string) error {
	filePath, err := FileWithPatternPath(dir, pattern)
	if err != nil {
		return err
	}
	if filePath == "" {
		return errors.New("Couldn't find file with pattern " + pattern + " in directory " + dir)
	} else {
		if err := os.Remove(filePath); err != nil {
			return err
		}
	}
	return nil
}

func getLightNodeKeysDir(root string) string {
	return filepath.Join(root, consts.ConfigDirName.DALightNode, consts.KeysDirName)
}

func VerifyCelestiaLightNodeKeys(root string) error {
	lightNodeKeysDir := filepath.Join(getLightNodeKeysDir(root), innerKeysDirName)

	infoFilePath := filepath.Join(lightNodeKeysDir, "my_celes_key"+".info")
	err := verifyFileExists(infoFilePath)
	if err != nil {
		return err
	}
	return verifyAndRemoveFilePattern(addressPattern, lightNodeKeysDir)
}

func getRelayerKeysDir(root string) string {
	return filepath.Join(root, consts.ConfigDirName.Relayer, consts.KeysDirName)
}

func VerifyRelayerKeys(root string, rollappID string, hubID string) error {
	relayerKeysDir := getRelayerKeysDir(root)
	rollappKeysDir := filepath.Join(relayerKeysDir, rollappID, innerKeysDirName)
	rollappKeyInfoPath := filepath.Join(rollappKeysDir, consts.KeysIds.RollappRelayer+".info")
	if err := verifyFileExists(rollappKeyInfoPath); err != nil {
		return err
	}
	if err := verifyAndRemoveFilePattern(addressPattern, rollappKeysDir); err != nil {
		return err
	}
	hubKeysDir := filepath.Join(relayerKeysDir, hubID, innerKeysDirName)
	hubKeyInfoPath := filepath.Join(hubKeysDir, consts.KeysIds.HubRelayer+".info")
	if err := verifyFileExists(hubKeyInfoPath); err != nil {
		return err
	}
	return verifyAndRemoveFilePattern(addressPattern, hubKeysDir)
}
