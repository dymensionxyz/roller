package dependencies

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
)

const (
	solcVersion = "0.8.20" // Latest stable version as of @20250121
)

// getSolcBinaryName returns the appropriate solc binary name based on the OS
func getSolcBinaryName() string {
	if runtime.GOOS == "darwin" {
		return "solc-macos"
	}
	return "solc-static-linux"
}

// InstallSolidityDependencies installs the solc binary for Solidity contract compilation
func InstallSolidityDependencies() error {
	// Create bin directory if it doesn't exist
	if err := os.MkdirAll(consts.InternalBinsDir, 0o755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	solcPath := filepath.Join(consts.InternalBinsDir, "solc")

	// Check if solc is already installed
	if _, err := os.Stat(solcPath); err == nil {
		// Already installed
		return nil
	}

	// Download solc based on OS
	binaryName := getSolcBinaryName()
	downloadURL := fmt.Sprintf(
		"https://github.com/ethereum/solidity/releases/download/v%s/%s",
		solcVersion,
		binaryName,
	)

	// Download the binary
	cmd := exec.Command("sudo", "curl", "-L", downloadURL, "-o", solcPath)
	if _, err := bash.ExecCommandWithStdout(cmd); err != nil {
		return fmt.Errorf("failed to download solc: %w", err)
	}

	// Make binary executable
	cmd = exec.Command("chmod", "+x", solcPath)
	if _, err := bash.ExecCommandWithStdout(cmd); err != nil {
		return fmt.Errorf("failed to make solc executable: %w", err)
	}

	return nil
}
