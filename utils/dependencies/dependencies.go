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
	"strconv"
	"strings"

	"github.com/pterm/pterm"
	"github.com/schollz/progressbar/v3"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/archives"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/dependencies/types"
	firebaseutils "github.com/dymensionxyz/roller/utils/firebase"
	genesisutils "github.com/dymensionxyz/roller/utils/genesis"
	"github.com/dymensionxyz/roller/utils/rollapp"
)

func InstallBinaries(withMockDA bool, raResp rollapp.ShowRollappResponse, env string) (
	map[string]types.Dependency,
	map[string]types.Dependency,
	error,
) {
	c := exec.Command("sudo", "mkdir", "-p", consts.InternalBinsDir)
	_, err := bash.ExecCommandWithStdout(c)
	if err != nil {
		errMsg := fmt.Sprintf("failed to create %s\n", consts.InternalBinsDir)
		return nil, nil, errors.New(errMsg)
	}

	genesisTmpDir, err := os.MkdirTemp(os.TempDir(), "genesis-file")
	if err != nil {
		return nil, nil, err
	}
	// nolint: errcheck
	defer os.RemoveAll(genesisTmpDir)

	var raCommit string
	var drsVersion string
	var da string

	raVmType := strings.ToLower(raResp.Rollapp.VmType)
	if !withMockDA {
		// TODO refactor, this genesis file fetch is redundand and will slow the process down
		// when the genesis file is big
		err = genesisutils.DownloadGenesis(genesisTmpDir, raResp.Rollapp.Metadata.GenesisUrl)
		if err != nil {
			pterm.Error.Println("failed to download genesis file: ", err)
			return nil, nil, err
		}

		as, err := genesisutils.GetGenesisAppState(genesisTmpDir)
		if err != nil {
			return nil, nil, err
		}

		da = as.RollappParams.Params.Da
		drsVersion = strconv.Itoa(as.RollappParams.Params.DrsVersion)
		pterm.Info.Println("RollApp drs version from the genesis file : ", drsVersion)
		drsInfo, err := firebaseutils.GetLatestDrsVersionCommit(drsVersion, env)
		if err != nil {
			return nil, nil, err
		}

		switch strings.ToLower(raResp.Rollapp.VmType) {
		case "evm":
			raCommit = drsInfo.EvmCommit
		case "wasm":
			raCommit = drsInfo.WasmCommit
		}

		if raCommit == "UNRELEASED" {
			return nil, nil, errors.New("rollapp does not support drs version: " + drsVersion)
		}

		pterm.Info.Println(
			"Latest RollApp binary commit for the current DRS version: ",
			raCommit[:6],
		)
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
		rbi := NewRollappBinaryInfo(
			raResp.Rollapp.GenesisInfo.Bech32Prefix,
			raCommit,
			raVmType,
		)

		buildableDeps = DefaultRollappBuildableDependencies(rbi, da)
	}

	goreleaserDeps := map[string]types.Dependency{}

	if !withMockDA && da == "celestia" {
		goreleaserDeps = DefaultCelestiaPrebuiltDependencies()
	}

	if withMockDA {
		// @20240913 libwasm is necessary on the host VM to be able to run the prebuilt rollapp binary
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
			return nil, nil, errors.New("unsupported OS")
		}

		downloadPath := fmt.Sprintf(
			"https://github.com/CosmWasm/wasmvm/releases/download/%s/%s",
			libVersion,
			libName,
		)

		fsc := exec.Command("sudo", "mkdir", "-p", outputPath)
		_, err := bash.ExecCommandWithStdout(fsc)
		if err != nil {
			return nil, nil, err
		}

		c := exec.Command("sudo", "wget", "-O", filepath.Join(outputPath, libName), downloadPath)
		_, err = bash.ExecCommandWithStdout(c)
		if err != nil {
			return nil, nil, err
		}

		if raVmType == "evm" {
			goreleaserDeps["rollapp"] = types.Dependency{
				DependencyName:  "rollapp-evm",
				RepositoryOwner: "artemijspavlovs",
				RepositoryName:  "rollapp-evm",
				RepositoryUrl:   "https://github.com/artemijspavlovs/rollapp-evm",
				Release:         "v2.3.4-pg-roller-02",
				Binaries: []types.BinaryPathPair{
					{
						Binary:            "rollappd",
						BinaryDestination: consts.Executables.RollappEVM,
					},
				},
			}
		} else if raVmType == "wasm" {
			goreleaserDeps["rollapp"] = types.Dependency{
				DependencyName:  "rollapp-wasm",
				RepositoryOwner: "artemijspavlovs",
				RepositoryName:  "rollapp-wasm",
				RepositoryUrl:   "https://github.com/artemijspavlovs/rollapp-wasm",
				Release:         "v1.0.0-rc04-roller-07",
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
			return nil, nil, errors.New(errMsg)
		}

	}

	for k, dep := range buildableDeps {
		err := InstallBinaryFromRepo(dep, k)
		if err != nil {
			errMsg := fmt.Sprintf("failed to build binary %s: %v", k, err)
			return nil, nil, errors.New(errMsg)
		}
	}

	return buildableDeps, goreleaserDeps, nil
}

func InstallBinaryFromRepo(dep types.Dependency, td string) error {
	spinner, _ := pterm.DefaultSpinner.Start(
		fmt.Sprintf("[%s] installing", dep.DependencyName),
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

	spinner.UpdateText(
		fmt.Sprintf("[%s] cloning the repository", dep.DependencyName),
	)
	c := exec.Command("git", "clone", dep.RepositoryUrl, targetDir)
	_, err = bash.ExecCommandWithStdout(c)
	if err != nil {
		spinner.Fail(
			fmt.Sprintf("[%s] failed to clone", dep.DependencyName),
		)

		return err
	}

	// Change directory to the cloned repo
	if err := os.Chdir(targetDir); err != nil {
		spinner.Fail(
			fmt.Sprintf("[%s] failed to create a temp directory", dep.DependencyName),
		)
		return err
	}

	if dep.Release != "main" {
		spinner.UpdateText(fmt.Sprintf("[%s] checking out %s", dep.DependencyName, dep.Release))
		if err := exec.Command("git", "checkout", dep.Release).Run(); err != nil {
			spinner.Fail(
				fmt.Sprintf("[%s] failed to checkout %s", dep.DependencyName, dep.Release),
			)
			return err
		}
	}

	var buildSource string
	if strings.HasPrefix(dep.Release, "v") {
		buildSource = dep.Release
	} else {
		buildSource = dep.Release[:6]
	}

	spinner.UpdateText(
		fmt.Sprintf(
			"[%s] starting build from %s (this can take several minutes)",
			dep.DependencyName,
			buildSource,
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
			fmt.Sprintf("[%s] finishing installation", filepath.Base(binary.BinaryDestination)),
		)
	}

	spinner.Success(fmt.Sprintf("[%s] installed\n", dep.DependencyName))
	return nil
}

func InstallBinaryFromRelease(dep types.Dependency) error {
	spinner, _ := pterm.DefaultSpinner.Start(
		fmt.Sprintf("[%s] installing", dep.DependencyName),
	)

	goOsCaser := cases.Title(language.Und)
	goOs := goOsCaser.String(runtime.GOOS)
	goArch := strings.ToLower(runtime.GOARCH)
	if goArch == "amd64" && dep.DependencyName == "celestia-app" {
		goArch = "x86_64"
	}

	if goArch == "amd64" && dep.DependencyName == "celestia-node" {
		goArch = "x86_64"
	}

	targetDir, err := os.MkdirTemp(os.TempDir(), dep.DependencyName)
	if err != nil {
		// nolint: errcheck,gosec
		return err
	}
	archiveName := fmt.Sprintf(
		"%s_%s_%s.tar.gz",
		dep.DependencyName,
		goOs,
		goArch,
	)
	// nolint: errcheck
	defer os.RemoveAll(targetDir)

	url := fmt.Sprintf(
		"%s/releases/download/%s/%s",
		dep.RepositoryUrl,
		dep.Release,
		archiveName,
	)

	spinner.UpdateText(fmt.Sprintf("[%s] downloading %s", dep.DependencyName, dep.Release))
	err = DownloadRelease(url, targetDir, dep, spinner)
	if err != nil {
		// nolint: errcheck,gosec
		spinner.Fail("failed to download release")
		return err
	}
	spinner.UpdateText(fmt.Sprintf("[%s] downloaded successfully", dep.DependencyName))

	spinner.Success(fmt.Sprintf("[%s] installed\n", dep.DependencyName))
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
	// nolint: errcheck
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

func DownloadBinary(url, destination string) error {
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

	// Create a temporary file
	f, err := os.CreateTemp("", "binary-")
	if err != nil {
		return err
	}
	tempName := f.Name()
	defer os.Remove(tempName) // Clean up in case of failure

	// Create a progress bar
	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"Downloading",
	)

	// Copy the response body to the temporary file while updating the progress bar
	_, err = io.Copy(io.MultiWriter(f, bar), resp.Body)
	if err != nil {
		return err
	}

	// Important: Close the file handle before moving it
	if err = f.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %w", err)
	}

	// Move the file into place and make it executable
	err = archives.MoveBinaryIntoPlaceAndMakeExecutable(tempName, destination)
	if err != nil {
		return err
	}

	return nil
}
