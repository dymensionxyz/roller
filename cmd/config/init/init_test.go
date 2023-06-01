package init

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestInitCmd(t *testing.T) {
	assert := assert.New(t)

	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "test")
	assert.NoError(err)

	// Cleanup after test finishes
	defer os.RemoveAll(tempDir)

	// Define flags
	hubRPC := "yourHubRPC"
	lightNodeEndpoint := "yourLightNodeEndpoint"
	keyPrefix := "yourKeyPrefix"
	rollappBinary := "yourRollappBinary"
	decimals := uint64(18)

	// Build the command
	cmd := InitCmd()
	cmd.SetArgs([]string{
		"chainID",
		"denom",
		"--" + flagNames.HubRPC, hubRPC,
		"--" + flagNames.LightNodeEndpoint, lightNodeEndpoint,
		"--" + flagNames.KeyPrefix, keyPrefix,
		"--" + flagNames.RollappBinary, rollappBinary,
		"--" + flagNames.Decimals, string(decimals),
	})

	// Run the command
	err = cmd.Execute()
	assert.NoError(err)

	// Check the on-disk output
	// TODO: Read the output file and assert its content.
	// This will depend on your specific use case.
	// The logic may look something like:
	// content, err := ioutil.ReadFile(filepath.Join(tempDir, "output.txt"))
	// assert.NoError(err)
	// assert.Equal("Expected content", string(content))
}