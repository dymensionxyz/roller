package filesystem

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/dustin/go-humanize"
	"github.com/hashicorp/go-extract"
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
	file, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %v", err)
	}
	// nolint:errcheck
	defer file.Close()

	const promptThreshold = 200 * 1024 * 1024 * 1024 // 200GB
	maxSize := int64(promptThreshold)

	// Check file size and prompt user if it exceeds threshold
	fileInfo, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to stat archive: %v", err)
	}

	if fileInfo.Size() > promptThreshold {
		pterm.Warning.Printf(
			"Archive size is %s, which exceeds the default %s limit\n",
			humanize.Bytes(uint64(fileInfo.Size())),
			humanize.Bytes(uint64(promptThreshold)),
		)
		pterm.Info.Println("Extraction may take significant time and disk space")

		proceed, _ := pterm.DefaultInteractiveConfirm.WithDefaultValue(false).
			WithDefaultText("Do you want to proceed with extraction?").Show()
		if !proceed {
			return fmt.Errorf("extraction cancelled by user")
		}

		maxSize = -1 // Disable limit if user approves
	}

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	spinner, _ := pterm.DefaultSpinner.Start("Extracting archive into " + destDir)
	// nolint:errcheck
	defer spinner.Stop()

	cfg := extract.NewConfig(
		extract.WithDenySymlinkExtraction(true),
		extract.WithMaxExtractionSize(maxSize),
		extract.WithMaxInputSize(maxSize),
	)

	ctx := context.Background()
	if err := extract.Unpack(ctx, destDir, file, cfg); err != nil {
		return fmt.Errorf("extraction failed: %v", err)
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
