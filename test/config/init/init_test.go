package initconfig_testing

import (
	"io/ioutil"
	"testing"

	"os"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	keys "github.com/dymensionxyz/roller/test/config/init/keys"
	utils "github.com/dymensionxyz/roller/test/config/init/utils"
	"github.com/stretchr/testify/assert"
)

func TestInitCmdWithoutParams(t *testing.T) {
	assert := assert.New(t)
	tempDir, err := ioutil.TempDir(os.TempDir(), "test")
	assert.NoError(err)
	defer func() {
		err := os.RemoveAll(tempDir)
		assert.NoError(err)
	}()
	cmd := initconfig.InitCmd()
	rollappID := "mars"
	cmd.SetArgs([]string{
		rollappID,
		"udym",
		"--" + initconfig.FlagNames.Home, tempDir,
	})
	assert.NoError(cmd.Execute())
	assert.NoError(keys.VerifyAllKeys(tempDir, rollappID, initconfig.DefaultHubID))
	assert.NoError(keys.ClearKeys(tempDir))
	are_dirs_equal, err := utils.CompareDirs(tempDir, "./goldens/init_without_flags")
	assert.NoError(err)
	assert.True(are_dirs_equal)
}

func TestInitCmdWithParams(t *testing.T) {
	assert := assert.New(t)
	tempDir, err := ioutil.TempDir(os.TempDir(), "test")
	assert.NoError(err)
	defer func() {
		err := os.RemoveAll(tempDir)
		assert.NoError(err)
	}()

	cmd := initconfig.InitCmd()
	rollappID := "mars"
	lightNodeEndpoint := "http://localhost:26659"
	denom := "udym"
	decimals := "6"

	cmd.SetArgs([]string{
		rollappID,
		denom,
		"--" + initconfig.FlagNames.Home, tempDir,
		"--" + initconfig.FlagNames.LightNodeEndpoint, lightNodeEndpoint,
		"--" + initconfig.FlagNames.Decimals, decimals})
	assert.NoError(cmd.Execute())

	assert.NoError(keys.VerifyRollappKeys(tempDir))
	assert.NoError(keys.VerifyRelayerKeys(tempDir, rollappID, initconfig.DefaultHubID))
	assert.NoError(keys.ClearKeys(tempDir))
	are_dirs_equal, err := utils.CompareDirs(tempDir, "./goldens/init_with_flags")
	assert.NoError(err)
	assert.True(are_dirs_equal)
}
