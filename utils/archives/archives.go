package archives

import (
	"archive/tar"
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
)

// ExtractZip function extracts the zip file created by the genesis-generator
// into a temporary directory and passes the tar archive container within
// the zip archive to ExtractTar for processing
func ExtractZip(zipFile string) error {
	tmpDir, err := os.MkdirTemp("", "genesis_zip_files")
	if err != nil {
		return err
	}
	// nolint errcheck
	defer os.RemoveAll(tmpDir)

	zipReader, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	// nolint errcheck
	defer zipReader.Close()

	var tarFilePath string

	// Iterate through the files in the ZIP archive
	for _, f := range zipReader.File {
		if filepath.Ext(f.Name) == ".tar" {
			// nolint gosec
			tarFilePath = filepath.Join(tmpDir, f.Name)
			if err := extractFileFromZip(f, tarFilePath); err != nil {
				return fmt.Errorf("failed to extract .tar file %s: %w", tarFilePath, err)
			}
		}
	}

	if tarFilePath == "" {
		return fmt.Errorf("no .tar file found in the zip archive")
	}

	// Process the extracted .tar file
	if err := ExtractTar(tarFilePath, tmpDir); err != nil {
		return fmt.Errorf("failed to extract .tar file %s: %w", tarFilePath, err)
	}

	return nil
}

// ExtractTar function extracts the tar archive created by the genesis-generator
// and moves the files into the correct location
func ExtractTar(tarFile, outputDir string) error {
	supportedFiles := []string{"roller.toml", "genesis.json"}

	// Open the .tar file
	file, err := os.Open(tarFile)
	if err != nil {
		return fmt.Errorf("failed to open tar file: %w", err)
	}
	// nolint errcheck
	defer file.Close()

	// Create a tar reader
	tarReader := tar.NewReader(file)

	// Iterate through the files in the tar archive
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return fmt.Errorf("ExtractTar: Next() failed: %w", err)
		}

		if !slices.Contains(supportedFiles, header.Name) {
			continue
		}

		switch header.Name {
		case "roller.toml":
			// Create directory
			filePath := filepath.Join(utils.GetRollerRootDir(), "roller.toml")
			err := createFileFromArchive(filePath, tarReader)
			if err != nil {
				return err
			}
		case "genesis.json":
			filePath := filepath.Join(
				utils.GetRollerRootDir(),
				consts.ConfigDirName.Rollapp,
				"config",
				"genesis.json",
			)
			err := createFileFromArchive(filePath, tarReader)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func createFileFromArchive(outputPath string, tarReader *tar.Reader) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("ExtractTar: MkdirAll() failed: %w", err)
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("ExtractTar: Create() failed: %w", err)
	}

	if _, err := io.Copy(outFile, tarReader); err != nil {
		// nolint errcheck
		outFile.Close()
		return fmt.Errorf("ExtractTar: Copy() failed: %w", err)
	}
	// nolint errcheck
	outFile.Close()
	return nil
}

// extractFileFromZip extracts a file from a zip archive to the specified path
func extractFileFromZip(f *zip.File, outputPath string) error {
	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("failed to open file in zip: %w", err)
	}
	// nolint errcheck
	defer rc.Close()

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	// nolint errcheck
	defer outFile.Close()

	// nolint gosec
	if _, err := io.Copy(outFile, rc); err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	return nil
}
