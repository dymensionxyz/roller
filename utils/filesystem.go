package utils

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/user"
	"path/filepath"

	"github.com/pterm/pterm"
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
func DownloadFile(url, filepath string) error {
	spinner, _ := pterm.DefaultSpinner.
		Start("Downloading file file from ", url)

	// nolint:gosec
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		// nolint:errcheck
		resp.Body.Close()
		spinner.Fail("failed to download file: ", err)
		return err
	}
	// nolint:errcheck
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		spinner.Fail("failed to download file: ", err)
		return err
	}
	// nolint:errcheck
	defer out.Close()

	spinner.Success("Successfully downloaded the genesis file")
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
