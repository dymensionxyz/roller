package initconfig_test

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"os"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/test/config/init/testutils"
	"github.com/stretchr/testify/assert"
)

func TestInitCmd(t *testing.T) {
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
				"--" + initconfig.FlagNames.DAEndpoint, "http://localhost:26659",
				"--" + initconfig.FlagNames.Decimals, "6",
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
			cmd := initconfig.InitCmd()
			rollappID := "mars"
			cmd.SetArgs(append([]string{
				rollappID,
				"udym",
				"--" + initconfig.FlagNames.Home, tempDir,
			}, tc.optionalFlags...))
			assert.NoError(cmd.Execute())
			assert.NoError(testutils.VerifyRollappKeys(tempDir))
			assert.NoError(testutils.VerifyRelayerKeys(tempDir, rollappID, initconfig.HubData.ID))
			if !testutils.Contains(tc.optionalFlags, "--"+initconfig.FlagNames.DAEndpoint) {
				assert.NoError(testutils.VerifyLightNodeKeys(tempDir))
			}
			assert.NoError(testutils.ClearKeys(tempDir))
			are_dirs_equal, err := testutils.CompareDirs(tempDir, tc.goldenDirPath)
			assert.NoError(err)
			assert.True(are_dirs_equal)
		})
	}
}
