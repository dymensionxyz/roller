package initconfig_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"github.com/dymensionxyz/roller/test/config/init/testutils"
)

func TestInitCmd(t *testing.T) {
	tokenSupply := "10000000"
	testCases := []struct {
		name          string
		goldenDirPath string
		excludedDirs  []string
		optionalFlags []string
	}{
		{
			name:          "Roller config init with default values",
			goldenDirPath: "./goldens/init_without_flags",
			excludedDirs:  []string{"gentx"},
			optionalFlags: []string{
				"--" + initconfig.FlagNames.HubID, "local",
			},
		},
		{
			name:          "Roller config init with custom flags",
			goldenDirPath: "./goldens/init_with_flags",
			excludedDirs:  []string{"gentx"},
			optionalFlags: []string{
				"--" + initconfig.FlagNames.TokenSupply, tokenSupply,
				"--" + initconfig.FlagNames.HubID, "local",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			tempDir, err := os.MkdirTemp(os.TempDir(), "test")
			tempDir = filepath.Join(tempDir, ".roller")
			fmt.Println(tempDir, tc.name)
			assert.NoError(err)
			defer func() {
				err := os.RemoveAll(tempDir)
				assert.NoError(err)
			}()
			initCmd := initconfig.InitCmd()
			utils.AddGlobalFlags(initCmd)
			denom := "dym"
			rollappID := "mars"
			initCmd.SetArgs(append([]string{
				rollappID,
				denom,
				"--" + utils.FlagNames.Home, tempDir,
			}, tc.optionalFlags...))
			assert.NoError(initCmd.Execute())
			rlpCfg, err := config.LoadConfigFromTOML(tempDir)
			assert.NoError(err)
			assert.NoError(os.Remove(filepath.Join(tempDir, config.RollerConfigFileName)))
			assert.NoError(testutils.VerifyRollappKeys(tempDir))
			assert.NoError(
				testutils.VerifyRelayerKeys(tempDir, rlpCfg.RollappID, rlpCfg.HubData.ID),
			)
			assert.NoError(testutils.VerifyCelestiaLightNodeKeys(tempDir))
			assert.NoError(testutils.SanitizeConfigDir(tempDir, &rlpCfg))
			assert.NoError(testutils.VerifyRlyConfig(rlpCfg, tc.goldenDirPath))
		})
	}
}
