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

	destDir, err := filepath.Abs(destDir)
	if err != nil {
		return fmt.Errorf("failed to resolve destination directory: %v", err)
	}

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

	const maxSize = 100 * 1024 * 1024 * 1024
	var totalSize int64

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar reading error: %v", err)
		}

		if header.Name != "data" && !filepath.HasPrefix(header.Name, "data/") {
			continue
		}

		cleanName := filepath.Clean(header.Name)
		if filepath.IsAbs(cleanName) {
			return fmt.Errorf("invalid path in archive (absolute path): %s", header.Name)
		}

		target := filepath.Join(destDir, cleanName)

		absTarget, err := filepath.Abs(target)
		if err != nil {
			return fmt.Errorf("failed to resolve target path: %v", err)
		}

		if !filepath.HasPrefix(absTarget, destDir) {
			return fmt.Errorf("invalid path in archive (escapes destination): %s", header.Name)
		}

		totalSize += header.Size
		if totalSize > maxSize {
			return fmt.Errorf("archive too large (exceeds %d bytes)", maxSize)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(absTarget, 0o755); err != nil {
				return fmt.Errorf("failed to create directory %s: %v", absTarget, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(absTarget), 0o755); err != nil {
				return fmt.Errorf("failed to create parent directory: %v", err)
			}

			f, err := os.OpenFile(absTarget, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o644)
			if err != nil {
				return fmt.Errorf("failed to create file %s: %v", absTarget, err)
			}

			written, err := io.CopyN(f, tr, header.Size)
			closeErr := f.Close()

			if err != nil && err != io.EOF {
				return fmt.Errorf("failed to write to file %s: %v", absTarget, err)
			}
			if closeErr != nil {
				return fmt.Errorf("failed to close file %s: %v", absTarget, closeErr)
			}
			if written != header.Size {
				return fmt.Errorf("file size mismatch for %s: expected %d, got %d", absTarget, header.Size, written)
			}
		case tar.TypeSymlink, tar.TypeLink:
			return fmt.Errorf("symlinks and hardlinks not allowed in archive: %s", header.Name)
		default:
			continue
		}
	}

	spinner.Success("Archive extracted successfully")
	return nil
}

func CompressTarGz(sourceDir, destDir, fileName string) error {
	spinner, _ := pterm.DefaultSpinner.Start("Creating archive from...")
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
		spinner.Fail("Archive compressed failed")
	} else {
		spinner.Success("Archive compressed successfully")
	}

	return nil
}
