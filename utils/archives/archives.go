package archives

import (
	"archive/tar"
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// TODO: work with actual gzip archive once genesis creator exports one

// TraverseTARFile function extracts a .tar file from a .zip archive and
// extracts the fileName into outputDir
func ExtractFileFromNestedTar(sourceZipFilePath, fileName, outputDir string) error {
	tmpDir, err := os.MkdirTemp("", "genesis_zip_files")
	if err != nil {
		return err
	}
	// nolint errcheck
	defer os.RemoveAll(tmpDir)

	zipReader, err := zip.OpenReader(sourceZipFilePath)
	if err != nil {
		return err
	}
	// nolint errcheck
	defer zipReader.Close()

	var tarFilePath string

	for _, f := range zipReader.File {
		if filepath.Ext(f.Name) == ".tar" {
			// nolint gosec
			tarFilePath = filepath.Join(tmpDir, f.Name)
			if err := extractFileFromZip(f, tarFilePath); err != nil {
				return fmt.Errorf("failed to extract .tar file %s: %w", tarFilePath, err)
			}

			err := TraverseTARFile(tarFilePath, fileName, outputDir)
			if err != nil {
				return fmt.Errorf("failed to traverse the tar file: %v ", err)
			}
		}
	}

	if tarFilePath == "" {
		return fmt.Errorf("no .tar file found in the zip archive")
	}

	return nil
}

// TraverseTARFile function traverses a .tar archuve and extracts the fileName into
// outputDir
func TraverseTARFile(tarFile, fileName, outputDir string) error {
	file, err := os.Open(tarFile)
	if err != nil {
		return fmt.Errorf("failed to open tar file: %w", err)
	}
	// nolint errcheck
	defer file.Close()
	tarReader := tar.NewReader(file)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return fmt.Errorf("ExtractTar: Next() failed: %w", err)
		}

		if fileName != header.Name {
			continue
		}

		if fileName == header.Name {
			fp := filepath.Join(outputDir, fileName)

			err := createFileFromArchive(fp, tarReader)
			if err != nil {
				return err
			}

			_, err = os.Stat(fp)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func createFileFromArchive(outputPath string, tarReader *tar.Reader) error {
	dir := filepath.Dir(outputPath)

	if err := os.MkdirAll(dir, 0o755); err != nil {
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
