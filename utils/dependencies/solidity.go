package dependencies

import (
	"os/exec"
	"runtime"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/dependencies/types"
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
	b := getSolcBinaryName()

	// Create bin directory if it doesn't exist
	dep := types.Dependency{
		DependencyName: "solc",
		RepositoryUrl:  "https://github.com/ethereum/solidity",
		RepositoryName: "solidity",
		Release:        solcVersion,
		Binaries: []types.BinaryPathPair{
			{
				Binary:            b,
				BinaryDestination: consts.Executables.Solc,
				BuildCommand:      exec.Command("make", "build"),
			},
		},
	}

	if err := InstallBinaryFromRelease(dep); err != nil {
		return err
	}

	return nil
}
