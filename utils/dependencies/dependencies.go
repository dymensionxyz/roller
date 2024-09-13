package dependencies

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/archives"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/dependencies/types"
)

func InstallBinaries(bech32 string, withMockDA bool) error {
	c := exec.Command("sudo", "mkdir", "-p", consts.InternalBinsDir)
	_, err := bash.ExecCommandWithStdout(c)
	if err != nil {
		errMsg := fmt.Sprintf("failed to create %s", consts.InternalBinsDir)
		return errors.New(errMsg)
	}

	defer func() {
		dir, err := os.UserHomeDir()
		if err != nil {
			return
		}
		_ = os.Chdir(dir)
	}()

	buildableDeps := map[string]types.Dependency{
		"rollapp": {
			Name:       "rollapp",
			Repository: "https://github.com/dymensionxyz/rollapp-evm.git",
			Release:    "e68f8190f1301b317846623a9e83be7acc2ad56e", // 20240909 rolapparams module
			Binaries: []types.BinaryPathPair{
				{
					Binary:            "./build/rollapp-evm",
					BinaryDestination: consts.Executables.RollappEVM,
					BuildCommand: exec.Command(
						"make",
						"build",
						fmt.Sprintf("BECH32_PREFIX=%s", bech32),
					),
				},
			},
		},
	}

	if !withMockDA {
		buildableDeps["celestia"] = types.Dependency{
			Name:       "celestia",
			Repository: "https://github.com/celestiaorg/celestia-node.git",
			Release:    "v0.16.0",
			Binaries: []types.BinaryPathPair{
				{
					Binary:            "./build/celestia",
					BinaryDestination: consts.Executables.Celestia,
					BuildCommand: exec.Command(
						"make",
						"build",
					),
				},
				{
					Binary:            "./cel-key",
					BinaryDestination: consts.Executables.CelKey,
					BuildCommand: exec.Command(
						"make",
						"cel-key",
					),
				},
			},
		}
	}

	goreleaserDeps := map[string]types.Dependency{
		"celestia-app": {
			Name:       "celestia-app",
			Repository: "https://github.com/celestiaorg/celestia-app",
			Release:    "2.1.2",
			Binaries: []types.BinaryPathPair{
				{
					Binary:            "celestia-appd",
					BinaryDestination: consts.Executables.CelestiaApp,
					BuildCommand: exec.Command(
						"make",
						"build",
					),
				},
			},
		},
		"eibc-client": {
			Name:       "eibc-client",
			Repository: "https://github.com/artemijspavlovs/eibc-client",
			Release:    "1.1.0",
			Binaries: []types.BinaryPathPair{
				{
					Binary:            "eibc-client",
					BinaryDestination: consts.Executables.Eibc,
				},
			},
		},
		"rly": {
			Name:       "go-relayer",
			Repository: "https://github.com/artemijspavlovs/go-relayer",
			Release:    "0.3.4-v2.5.2-relayer-canon-3",
			Binaries: []types.BinaryPathPair{
				{
					Binary:            "rly",
					BinaryDestination: consts.Executables.Relayer,
				},
			},
		},
	}
	//
	for k, dep := range goreleaserDeps {
		err := InstallBinaryFromRelease(dep)
		if err != nil {
			errMsg := fmt.Sprintf("failed to build binary %s: %v", k, err)
			return errors.New(errMsg)
		}

	}

	for k, dep := range buildableDeps {
		err := InstallBinaryFromRepo(dep, k)
		if err != nil {
			errMsg := fmt.Sprintf("failed to build binary %s: %v", k, err)
			return errors.New(errMsg)
		}
	}

	return nil
}

func InstallBinaryFromRepo(dep types.Dependency, td string) error {
	spinner, _ := pterm.DefaultSpinner.Start(fmt.Sprintf("Installing %s", dep.Name))
	targetDir, err := os.MkdirTemp(os.TempDir(), td)
	if err != nil {
		return err
	}
	// nolint: errcheck
	defer os.RemoveAll(targetDir)
	// Clone the repository
	err = os.Chdir(targetDir)
	if err != nil {
		pterm.Error.Println("failed to create a temp directory")
		return err
	}

	spinner.UpdateText(
		fmt.Sprintf("cloning %s into %s", dep.Repository, targetDir),
	)

	c := exec.Command("git", "clone", dep.Repository, targetDir)
	_, err = bash.ExecCommandWithStdout(c)
	if err != nil {
		pterm.Error.Println("failed to clone")
		return err
	}
	// Change directory to the cloned repo
	if err := os.Chdir(targetDir); err != nil {
		pterm.Error.Println("failed to create a temp directory")
		return err
	}

	if dep.Release != "main" {
		// Checkout a specific version (e.g., a tag or branch)
		spinner.UpdateText(fmt.Sprintf("checking out %s", dep.Release))
		if err := exec.Command("git", "checkout", dep.Release).Run(); err != nil {
			spinner.Fail(fmt.Sprintf("failed to checkout: %v\n", err))
			return err
		}
	}

	spinner.UpdateText(
		fmt.Sprintf(
			"starting %s build from %s (this can take several minutes)",
			dep.Name,
			dep.Release,
		),
	)

	// Build the binary
	for _, binary := range dep.Binaries {
		_, err := bash.ExecCommandWithStdout(binary.BuildCommand)
		spinner.UpdateText(fmt.Sprintf("building %s\n", binary.Binary))
		if err != nil {
			spinner.Fail(fmt.Sprintf("failed to build binary %s: %v\n", binary.BuildCommand, err))
			return err
		}

		c := exec.Command("sudo", "mv", binary.Binary, binary.BinaryDestination)
		if _, err := bash.ExecCommandWithStdout(c); err != nil {
			spinner.Fail(
				fmt.Sprintf(
					"Failed to move binary %s to %s\n",
					binary.Binary,
					binary.BinaryDestination,
				),
			)
			return err
		}
		spinner.Success(fmt.Sprintf("Successfully installed %s\n", binary.BinaryDestination))
	}

	spinner.Success(fmt.Sprintf("Successfully installed %s\n", dep.Name))
	return nil
}

func InstallBinaryFromRelease(dep types.Dependency) error {
	spinner, _ := pterm.DefaultSpinner.Start(fmt.Sprintf("Installing %s", dep.Name))
	goOs := strings.Title(runtime.GOOS)
	goArch := strings.ToLower(runtime.GOARCH)
	if goArch == "amd64" && dep.Name == "celestia-app" {
		goArch = "x86_64"
	}

	targetDir, err := os.MkdirTemp(os.TempDir(), dep.Name)
	if err != nil {
		// nolint: errcheck
		spinner.Stop()
		return err
	}
	archiveName := fmt.Sprintf(
		"%s_%s_%s.tar.gz",
		dep.Name,
		goOs,
		goArch,
	)
	// nolint: errcheck
	defer os.RemoveAll(targetDir)

	url := fmt.Sprintf(
		"%s/releases/download/v%s/%s",
		dep.Repository,
		dep.Release,
		archiveName,
	)

	err = DownloadRelease(url, targetDir, dep)
	if err != nil {
		// nolint: errcheck
		spinner.Stop()
		return err
	}

	spinner.Success(fmt.Sprintf("Successfully installed %s\n", dep.Name))
	return nil
}

func DownloadRelease(url, destination string, dep types.Dependency) error {
	// nolint gosec
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	// nolint errcheck
	defer resp.Body.Close()
	err = archives.ExtractTarGz(destination, resp.Body, dep)
	if err != nil {
		return err
	}

	return nil
}
