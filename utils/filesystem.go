package utils

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"

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

func DownloadArchive(url string) (io.ReadCloser, error) {
	// nolint:gosec
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		// nolint:errcheck
		resp.Body.Close()
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	return resp.Body, nil
}

func ExtractTarGz(gzipStream io.Reader, destDir string) error {
	spinner, _ := pterm.DefaultSpinner.Start("Extracting archive...")
	// nolint:errcheck
	defer spinner.Stop()

	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %v", err)
	}
	defer uncompressedStream.Close()

	tarReader := tar.NewReader(uncompressedStream)

	var foundDataDir bool
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar reading error: %v", err)
		}

		// Check if the archive contains a 'data' directory at its root
		if header.Name == "data/" || header.Name == "data" {
			foundDataDir = true
		}

		// Only process files within the 'data' directory
		if !strings.HasPrefix(header.Name, "data/") && header.Name != "data" {
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
			defer f.Close()

			// nolint:gosec
			if _, err := io.Copy(f, tarReader); err != nil {
				return fmt.Errorf("failed to write to file %s: %v", target, err)
			}
		}
	}

	if !foundDataDir {
		return fmt.Errorf("archive does not contain a 'data' directory at its root")
	}

	spinner.Success("Archive extracted successfully")
	return nil
}
