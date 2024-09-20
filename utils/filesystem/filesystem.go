package filesystem

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"

	"github.com/nxadm/tail"
	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/utils/bash"
)

func DirNotEmpty(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	if !info.IsDir() {
		return false, fmt.Errorf("%s is not a directory", path)
	}

	files, err := os.ReadDir(path)
	return len(files) > 0, err
}

func MoveFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer func() { _ = srcFile.Close() }()
	// nolint:gofumpt
	err = os.MkdirAll(filepath.Dir(dst), 0o750)
	if err != nil {
		return fmt.Errorf("failed to create parent directories: %w", err)
	}
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func() { _ = dstFile.Close() }()
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}
	err = os.Remove(src)
	if err != nil {
		return fmt.Errorf("failed to delete source file: %w", err)
	}
	return nil
}

func ExpandHomePath(path string) (string, error) {
	if path[:2] == "~/" {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		path = filepath.Join(usr.HomeDir, path[2:])
	}
	return path, nil
}

// TODO: download the file in chunks if possible
func DownloadFile(url, fp string) error {
	err := os.MkdirAll(filepath.Dir(fp), 0o755)
	if err != nil {
		return err
	}

	spinner, _ := pterm.DefaultSpinner.
		Start("Downloading ", filepath.Base(fp))
	fmt.Println()

	// nolint:gosec
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		// nolint:errcheck,gosec
		resp.Body.Close()
		spinner.Fail("failed to download file: ", err)
		return err
	}
	// nolint:errcheck
	defer resp.Body.Close()

	out, err := os.Create(fp)
	if err != nil {
		spinner.Fail("failed to download file: ", err)
		return err
	}
	// nolint:errcheck
	defer out.Close()

	spinner.Success(fmt.Sprintf("Successfully downloaded the %s", filepath.Base(fp)))
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func DownloadAndSaveArchive(url string, destPath string) (string, error) {
	spinner, _ := pterm.DefaultSpinner.Start("Downloading file...")

	// Create the destination directory if it doesn't exist
	err := os.MkdirAll(filepath.Dir(destPath), 0o755)
	if err != nil {
		spinner.Fail(fmt.Sprintf("Failed to create destination directory: %v", err))
		return "", fmt.Errorf("failed to create destination directory: %v", err)
	}

	// Download the file
	// nolint:gosec
	resp, err := http.Get(url)
	if err != nil {
		spinner.Fail(fmt.Sprintf("Failed to download file: %v", err))
		return "", fmt.Errorf("failed to download file: %v", err)
	}
	// nolint:errcheck
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		spinner.Fail(fmt.Sprintf("Bad status: %s", resp.Status))
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	// Create the destination file
	out, err := os.Create(destPath)
	if err != nil {
		spinner.Fail(fmt.Sprintf("Failed to create file: %v", err))
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	// nolint:errcheck
	defer out.Close()

	// Create a hash writer
	hash := sha256.New()
	writer := io.MultiWriter(out, hash)

	// Copy the body to file and hash
	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		spinner.Fail(fmt.Sprintf("Failed to save file: %v", err))
		return "", fmt.Errorf("failed to save file: %v", err)
	}

	hashStr := fmt.Sprintf("%x", hash.Sum(nil))
	spinner.Success("File downloaded and saved successfully")
	return hashStr, nil
}

func ExtractTarGz(sourcePath, destDir string) error {
	spinner, _ := pterm.DefaultSpinner.Start("Extracting archive...")
	// nolint:errcheck
	defer spinner.Stop()

	file, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %v", err)
	}
	// nolint:errcheck
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %v", err)
	}
	// nolint:errcheck
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar reading error: %v", err)
		}

		// Ensure we only extract the 'data' directory
		if header.Name != "data" && !filepath.HasPrefix(header.Name, "data/") {
			continue
		}

		// nolint:gosec
		target := filepath.Join(destDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return fmt.Errorf("failed to create directory %s: %v", target, err)
			}
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file %s: %v", target, err)
			}
			// nolint:errcheck
			defer f.Close()

			// nolint:gosec
			if _, err := io.Copy(f, tr); err != nil {
				return fmt.Errorf("failed to write to file %s: %v", target, err)
			}
		}
	}

	spinner.Success("Archive extracted successfully")
	return nil
}

func RemoveFileIfExists(filePath string) error {
	_, err := os.Stat(filePath)
	if err == nil {
		c := exec.Command("sudo", "rm", "-rf", filePath)
		_, err := bash.ExecCommandWithStdout(c)
		if err != nil {
			return fmt.Errorf("failed to remove file: %w", err)
		}
		fmt.Printf("File %s has been removed\n", filePath)
	} else if os.IsNotExist(err) {
		fmt.Printf("File %s does not exist\n", filePath)
	} else {
		return fmt.Errorf("error checking file: %w", err)
	}
	return nil
}

func TailFile(fp, svcName string) error {
	t, err := tail.TailFile(fp, tail.Config{Follow: true, ReOpen: false})
	if err != nil {
		return fmt.Errorf("failed to tail file: %v", err)
	}

	infoPrefix := pterm.Info.Prefix
	infoPrefix.Text = svcName
	cp := pterm.PrefixPrinter{
		Prefix: infoPrefix,
	}

	for line := range t.Lines {
		cp.Println(line.Text)
	}

	return nil
}

func UpdateHostsFile(addr, host string) error {
	// Check if the entry already exists
	pterm.Info.Printf("adding %s to hosts file\n", host)
	pterm.Debug.Println(
		"this is necessary to access the rollapp endpoint on your local machine from the docker" +
			" container",
	)
	checkCmd := exec.Command("grep", "-q", host, "/etc/hosts")
	err := checkCmd.Run()

	if err == nil {
		// Entry already exists
		fmt.Printf("Entry for %s already exists in /etc/hosts\n", host)
		return nil
	}

	// Append the new entry
	appendCmd := exec.Command(
		"sudo",
		"sh",
		"-c",
		fmt.Sprintf("echo '%s %s' >> /etc/hosts", addr, host),
	)
	output, err := appendCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to update hosts file: %v - %s", err, string(output))
	}

	fmt.Printf("Added %s to /etc/hosts\n", host)
	return nil
}
