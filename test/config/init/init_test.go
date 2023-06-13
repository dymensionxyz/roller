package initconfig_test

import (
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
	decimals := "6"
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
				"--" + initconfig.FlagNames.Decimals, decimals,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			tempDir, err := ioutil.TempDir(os.TempDir(), "test")
			tempDir = filepath.Join(tempDir, ".roller")
			assert.NoError(err)
			defer func() {
				err := os.RemoveAll(tempDir)
				assert.NoError(err)
			}()
			initCmd := initconfig.InitCmd()
			denom := "udym"
			rollappID := "mars"
			initCmd.SetArgs(append([]string{
				rollappID,
				denom,
				"--" + utils.FlagNames.Home, tempDir,
			}, tc.optionalFlags...))
			assert.NoError(initCmd.Execute())
			initConfig := initconfig.GetInitConfig(initCmd, []string{rollappID, denom})
			assert.NoError(testutils.VerifyRollerConfig(initConfig))
			assert.NoError(os.Remove(filepath.Join(tempDir, initconfig.RollerConfigFileName)))
			assert.NoError(testutils.VerifyRollappKeys(tempDir))
			assert.NoError(testutils.VerifyRelayerKeys(tempDir, rollappID, initConfig.HubData.ID))
			assert.NoError(testutils.VerifyLightNodeKeys(tempDir))
			assert.NoError(testutils.ClearKeys(tempDir))
			are_dirs_equal, err := testutils.CompareDirs(tempDir, tc.goldenDirPath)
			assert.NoError(err)
			assert.True(are_dirs_equal)
		})
	}
}
