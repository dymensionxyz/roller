package archives

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/dymensionxyz/roller/utils/bash"
	dependencytypes "github.com/dymensionxyz/roller/utils/dependencies/types"
)

// TODO: work with actual gzip archive once genesis creator exports one
func ExtractTarGz(path string, data io.ReadCloser, dep dependencytypes.Dependency) error {
	gzipReader, err := gzip.NewReader(data)
	if err != nil {
		log.Fatal(err)
	}
	defer gzipReader.Close()
	defer os.RemoveAll(path)

	tarReader := tar.NewReader(gzipReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			// nolint gosec
			if err := os.MkdirAll(filepath.Join(path, header.Name), 0o755); err != nil {
				log.Fatal(err)
			}
		case tar.TypeReg:
			// nolint gosec
			outFile, err := os.Create(filepath.Join(path, header.Name))
			if err != nil {
				log.Fatal(err)
			}
			// nolint gosec
			if _, err := io.Copy(outFile, tarReader); err != nil {
				log.Fatal(err)
			}
			outFile.Close()
		default:
			log.Printf("Unknown type: %x in %s\n", header.Typeflag, header.Name)
		}
	}

	for _, bin := range dep.Binaries {
		err := bash.ExecCommandWithInteractions(
			"sudo",
			"mv",
			filepath.Join(path, bin.Binary),
			bin.BinaryDestination,
		)
		if err != nil {
			return err
		}
		err = bash.ExecCommandWithInteractions(
			"sudo",
			"chmod", "+x",
			bin.BinaryDestination,
		)
		if err != nil {
			return err
		}
	}

	return err
}
