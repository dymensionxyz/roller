package filesystem

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pterm/pterm"
)

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
			// nolint: gosec
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, 0o755)
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

func CompressTarGz(sourceDir, destDir, fileName string) error {
	spinner, _ := pterm.DefaultSpinner.Start("Compressing archive...")
	// nolint:errcheck
	defer spinner.Stop()

	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create folder: %v", err)
	}

	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %v", err)
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		header.Name, _ = filepath.Rel(sourceDir, path)

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(tarWriter, file); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		spinner.Fail("Archive compressed successfully")
	} else {
		spinner.Success("Archive compressed successfully")
	}

	return nil
}
