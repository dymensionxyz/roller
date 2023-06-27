package testutils

import (
	"errors"
	"io/ioutil"

	"os"

	"github.com/dymensionxyz/roller/cmd/consts"
	"path/filepath"
	"regexp"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
)

const innerKeysDirName = "keyring-test"
const addressPattern = `.*\.address`

func SanitizeConfigDir(root string) error {
	keyDirs := []string{getLightNodeKeysDir(root), getRelayerKeysDir(root), getRollappKeysDir(root), getHubKeysDir(root)}
	for _, dir := range keyDirs {
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
	if err := SanitizeDymintToml(root); err != nil {
		return err
	}
	return nil
}

func SanitizeDymintToml(root string) error {
	dymintTomlPath := filepath.Join(root, consts.ConfigDirName.Rollapp, "config", "dymint.toml")
	return os.Remove(dymintTomlPath)
}
func verifyFileExists(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		return errors.New("File " + path + " does not exist")
	}
	return nil
}

func FileWithPatternPath(dir, pattern string) (string, error) {
	files, err := ioutil.ReadDir(dir)
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

func VerifyLightNodeKeys(root string) error {
	lightNodeKeysDir := filepath.Join(getLightNodeKeysDir(root), innerKeysDirName)
	infoFilePath := filepath.Join(lightNodeKeysDir, consts.KeyNames.DALightNode+".info")
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
	rollappKeyInfoPath := filepath.Join(rollappKeysDir, consts.KeyNames.RollappRelayer+".info")
	if err := verifyFileExists(rollappKeyInfoPath); err != nil {
		return err
	}
	if err := verifyAndRemoveFilePattern(addressPattern, rollappKeysDir); err != nil {
		return err
	}
	hubKeysDir := filepath.Join(relayerKeysDir, hubID, innerKeysDirName)
	hubKeyInfoPath := filepath.Join(hubKeysDir, consts.KeyNames.HubRelayer+".info")
	if err := verifyFileExists(hubKeyInfoPath); err != nil {
		return err
	}
	return verifyAndRemoveFilePattern(addressPattern, hubKeysDir)
}
