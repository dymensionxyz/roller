package dependencies

import (
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
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/dymensionxyz/roller/utils/archives"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/dependencies/types"
)

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
