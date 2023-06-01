package init

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"os"
)

func TestInitCmd(t *testing.T) {
	assert := assert.New(t)
	tempDir, err := ioutil.TempDir("", "test")
	defer func() {
		err := os.RemoveAll(tempDir)
		assert.NoError(err)
	}()
	assert.NoError(err)
	cmd := InitCmd()
	cmd.SetArgs([]string{
		"mars",
		"udym",
		"--" + flagNames.Home, tempDir,
	})
	err = cmd.Execute()
	assert.NoError(err)
}
