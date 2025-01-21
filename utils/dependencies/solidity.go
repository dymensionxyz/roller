package dependencies

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/pterm/pterm"

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

	if err := installSolcFromRelease(dep); err != nil {
		return err
	}

	return nil
}

// since solc doesn't follow the release artifact naming scheme, these
// are separate implementations for their binaries
func installSolcFromRelease(dep types.Dependency) error {
	spinner, _ := pterm.DefaultSpinner.Start(
		fmt.Sprintf("[%s] installing", dep.DependencyName),
	)

	url := fmt.Sprintf(
		"%s/releases/download/%s/%s",
		dep.RepositoryUrl,
		dep.Release,
		dep.Binaries[0].Binary,
	)

	spinner.UpdateText(fmt.Sprintf("[%s] downloading %s", dep.DependencyName, dep.Release))
	err := DownloadBinary(url, consts.Executables.Solc)
	if err != nil {
		spinner.Fail("failed to download release")
		return err
	}
	spinner.UpdateText(fmt.Sprintf("[%s] downloaded successfully", dep.DependencyName))

	spinner.Success(fmt.Sprintf("[%s] installed\n", dep.DependencyName))
	return nil
}
