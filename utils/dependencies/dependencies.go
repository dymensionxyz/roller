package dependencies

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pterm/pterm"
	"github.com/schollz/progressbar/v3"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/archives"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/dependencies/types"
	genesisutils "github.com/dymensionxyz/roller/utils/genesis"
	"github.com/dymensionxyz/roller/utils/rollapp"
	"github.com/dymensionxyz/roller/utils/roller"
)

func InstallBinaries(
	home string,
	withMockDA bool,
	raResp rollapp.ShowRollappResponse,
) error {
	c := exec.Command("sudo", "mkdir", "-p", consts.InternalBinsDir)
	_, err := bash.ExecCommandWithStdout(c)
	if err != nil {
		errMsg := fmt.Sprintf("failed to create %s\n", consts.InternalBinsDir)
		return errors.New(errMsg)
	}

	var raBinCommit string
	raVmType := strings.ToLower(raResp.Rollapp.VmType)
	raBech32Prefix := raResp.Rollapp.GenesisInfo.Bech32Prefix
	if !withMockDA {
		// TODO refactor, this genesis file fetch is redundand and will slow the process down
		// when the genesis file is big
		err = genesisutils.DownloadGenesis(home, raResp.Rollapp.Metadata.GenesisUrl)
		if err != nil {
			pterm.Error.Println("failed to download genesis file: ", err)
			return err
		}

		as, err := genesisutils.GetGenesisAppState(home)
		if err != nil {
			return err
		}

		err = os.RemoveAll(roller.GetRootDir())
		if err != nil {
			return err
		}
		raBinCommit = as.RollappParams.Params.Version
		pterm.Info.Println("RollApp binary version from the genesis file : ", raBinCommit)
	}

	defer func() {
		dir, err := os.UserHomeDir()
		if err != nil {
			return
		}
		_ = os.Chdir(dir)
	}()

	buildableDeps := map[string]types.Dependency{}

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

		if raVmType == "evm" {
			buildableDeps["rollapp"] = types.Dependency{
				Name:       "rollapp",
				Repository: "https://github.com/dymensionxyz/rollapp-evm.git",
				Release:    raBinCommit,
				Binaries: []types.BinaryPathPair{
					{
						Binary:            "./build/rollapp-evm",
						BinaryDestination: consts.Executables.RollappEVM,
						BuildCommand: exec.Command(
							"make",
							"build",
							fmt.Sprintf("BECH32_PREFIX=%s", raBech32Prefix),
						),
					},
				},
			}
		} else if raVmType == "wasm" {
			buildableDeps["rollapp"] = types.Dependency{
				Name:       "rollapp",
				Repository: "https://github.com/dymensionxyz/rollapp-wasm.git",
				Release:    raBinCommit,
				Binaries: []types.BinaryPathPair{
					{
						Binary:            "./build/rollapp-wasm",
						BinaryDestination: consts.Executables.RollappEVM,
						BuildCommand: exec.Command(
							"make",
							"build",
							fmt.Sprintf("BECH32_PREFIX=%s", raBech32Prefix),
						),
					},
				},
			}
		} else {
			return fmt.Errorf("RollApp VM '%s' type is not supported", raVmType)
		}
	}

	goreleaserDeps := map[string]types.Dependency{}

	if !withMockDA {
		necessaryDeps := map[string]types.Dependency{
			"celestia-app": {
				Name:       "celestia-app",
				Repository: "https://github.com/celestiaorg/celestia-app",
				Release:    "v2.1.2",
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
				Release:    "v1.1.4-roller",
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
				Release:    "v0.4.0-v2.5.2-relayer-pg-roller",
				Binaries: []types.BinaryPathPair{
					{
						Binary:            "rly",
						BinaryDestination: consts.Executables.Relayer,
					},
				},
			},
		}

		for s, dependency := range necessaryDeps {
			goreleaserDeps[s] = dependency
		}
	}

	if withMockDA {
		// @20240913 libwasm is necessary on the host VM to be able to run the rollapp binary
		var outputPath string
		var libName string
		libVersion := "v1.2.3"

		if runtime.GOOS == "linux" {
			outputPath = "/usr/lib"
			if runtime.GOARCH == "arm64" {
				libName = "libwasmvm.aarch64.so"
			} else if runtime.GOARCH == "amd64" {
				libName = "libwasmvm.x86_64.so"
			}
		} else if runtime.GOOS == "darwin" {
			outputPath = "/usr/local/lib"
			libName = "libwasmvm.dylib"
		} else {
			return errors.New("unsupported OS")
		}

		downloadPath := fmt.Sprintf(
			"https://github.com/CosmWasm/wasmvm/releases/download/%s/%s",
			libVersion,
			libName,
		)

		fsc := exec.Command("sudo", "mkdir", "-p", outputPath)
		_, err := bash.ExecCommandWithStdout(fsc)
		if err != nil {
			return err
		}

		c := exec.Command("sudo", "wget", "-O", filepath.Join(outputPath, libName), downloadPath)
		_, err = bash.ExecCommandWithStdout(c)
		if err != nil {
			return err
		}

		if raVmType == "evm" {
			goreleaserDeps["rollapp"] = types.Dependency{
				Name:       "rollapp-evm",
				Repository: "https://github.com/artemijspavlovs/rollapp-evm",
				Release:    "v2.3.4-pg-roller-02",
				Binaries: []types.BinaryPathPair{
					{
						Binary:            "rollappd",
						BinaryDestination: consts.Executables.RollappEVM,
					},
				},
			}
		} else if raVmType == "wasm" {
			goreleaserDeps["rollapp"] = types.Dependency{
				Name:       "rollapp-wasm",
				Repository: "https://github.com/artemijspavlovs/rollapp-wasm",
				Release:    "v1.0.0-rc04-roller-07",
				Binaries: []types.BinaryPathPair{
					{
						Binary:            "rollappd",
						BinaryDestination: consts.Executables.RollappEVM,
					},
				},
			}
		}

	}

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
	spinner, _ := pterm.DefaultSpinner.Start(
		fmt.Sprintf("Installing %s", dep.Name),
	)
	targetDir, err := os.MkdirTemp(os.TempDir(), td)
	if err != nil {
		return err
	}
	// nolint: errcheck
	defer os.RemoveAll(targetDir)
	// Clone the repository
	err = os.Chdir(targetDir)
	if err != nil {
		spinner.Fail("failed to create a temp directory")
		return err
	}

	c := exec.Command("git", "clone", dep.Repository, targetDir)
	_, err = bash.ExecCommandWithStdout(c)
	if err != nil {
		spinner.Fail("failed to clone")
		return err
	}
	// Change directory to the cloned repo
	if err := os.Chdir(targetDir); err != nil {
		spinner.Fail("failed to create a temp directory")
		return err
	}

	if dep.Release != "main" {
		// Checkout a specific version (e.g., a tag or branch)
		if err := exec.Command("git", "checkout", dep.Release).Run(); err != nil {
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
		if err != nil {
			spinner.Fail("failed to build")
			return err
		}

		c := exec.Command("sudo", "mv", binary.Binary, binary.BinaryDestination)
		if _, err := bash.ExecCommandWithStdout(c); err != nil {
			spinner.Fail("failed to install")
			return err
		}
		spinner.UpdateText(
			fmt.Sprintf("Finishing installation %s", filepath.Base(binary.BinaryDestination)),
		)
	}

	spinner.Success(fmt.Sprintf("Successfully installed %s\n", dep.Name))
	return nil
}

func InstallBinaryFromRelease(dep types.Dependency) error {
	spinner, _ := pterm.DefaultSpinner.Start(
		fmt.Sprintf("Installing %s", dep.Name),
	)
	goOs := strings.Title(runtime.GOOS)
	goArch := strings.ToLower(runtime.GOARCH)
	if goArch == "amd64" && dep.Name == "celestia-app" {
		goArch = "x86_64"
	}

	targetDir, err := os.MkdirTemp(os.TempDir(), dep.Name)
	if err != nil {
		// nolint: errcheck,gosec
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
		"%s/releases/download/%s/%s",
		dep.Repository,
		dep.Release,
		archiveName,
	)

	spinner.UpdateText(fmt.Sprintf("Downloading %s %s", dep.Name, dep.Release))
	err = DownloadRelease(url, targetDir, dep, spinner)
	if err != nil {
		// nolint: errcheck,gosec
		spinner.Fail("failed to download release")
		return err
	}
	spinner.UpdateText(fmt.Sprintf("Successfully downloaded %s", dep.Name))

	spinner.Success(fmt.Sprintf("Successfully installed %s\n", dep.Name))
	return nil
}

func DownloadRelease(
	url, destination string,
	dep types.Dependency,
	spinner *pterm.SpinnerPrinter,
) error {
	// Create a new HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create a progress bar
	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"Downloading",
	)

	// Create a reader that will update the progress bar
	reader := progressbar.NewReader(resp.Body, bar)

	// Create a pointer to the reader
	readerPtr := &reader

	// Create a wrapper that implements io.ReadCloser
	readCloserWrapper := struct {
		io.Reader
		io.Closer
	}{
		Reader: readerPtr,
		Closer: resp.Body,
	}

	// nolint: errcheck,gosec
	spinner.Stop()
	// Extract the tar.gz file with progress
	err = archives.ExtractTarGz(destination, readCloserWrapper, dep)
	if err != nil {
		return err
	}

	return nil
}
