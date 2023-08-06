package initconfig_test

import (
	"fmt"
	"github.com/dymensionxyz/roller/config"
	"io/ioutil"
	"path/filepath"
	"testing"

	"os"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/test/config/init/testutils"
	"github.com/stretchr/testify/assert"
)

func TestInitCmd(t *testing.T) {
	tokenSupply := "1000"
	testCases := []struct {
		name          string
		goldenDirPath string
		optionalFlags []string
	}{
		{
			name:          "Roller config init with default values",
			goldenDirPath: "./goldens/init_without_flags",
			optionalFlags: []string{},
		},
		{
			name:          "Roller config init with custom flags",
			goldenDirPath: "./goldens/init_with_flags",
			optionalFlags: []string{
				"--" + initconfig.FlagNames.TokenSupply, tokenSupply,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			tempDir, err := ioutil.TempDir(os.TempDir(), "test")
			tempDir = filepath.Join(tempDir, ".roller")
			fmt.Println(tc.name, tempDir)
			assert.NoError(err)
			//defer func() {
			//	err := os.RemoveAll(tempDir)
			//	assert.NoError(err)
			//}()
			initCmd := initconfig.InitCmd()
			utils.AddGlobalFlags(initCmd)
			denom := "dym"
			rollappID := "mars_1-1"
			initCmd.SetArgs(append([]string{
				rollappID,
				denom,
				"--" + utils.FlagNames.Home, tempDir,
			}, tc.optionalFlags...))
			assert.NoError(initCmd.Execute())
			initConfig, err := initconfig.GetInitConfig(initCmd, []string{rollappID, denom})
			assert.NoError(err)
			assert.NoError(testutils.VerifyRollerConfig(initConfig))
			assert.NoError(os.Remove(filepath.Join(tempDir, config.RollerConfigFileName)))
			assert.NoError(testutils.VerifyRollappKeys(tempDir))
			assert.NoError(testutils.VerifyRelayerKeys(tempDir, rollappID, initConfig.HubData.ID))
			assert.NoError(testutils.VerifyCelestiaLightNodeKeys(tempDir))
			assert.NoError(testutils.SanitizeConfigDir(tempDir))
			areDirsEqual, err := testutils.CompareDirs(tempDir, tc.goldenDirPath)
			assert.NoError(err)
			assert.True(areDirsEqual)
		})
	}
}
