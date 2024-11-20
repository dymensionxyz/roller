package templates

import (
	"embed"
	"html/template"
	"log"
	"os"
	"path/filepath"

	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/keys"
)

func WriteToFile(
	tmplPath, destFilePath string,
	ki keys.KeyInfo,
	embeddedTemplates embed.FS,
) error {
	tmpl := template.Must(template.ParseFS(embeddedTemplates, tmplPath))

	dir := filepath.Dir(destFilePath)
	ok, err := filesystem.DoesFileExist(dir)
	if !ok {
		err := os.MkdirAll(dir, 0o755)
		if err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}

	// Create or open the output file
	file, err := os.Create(destFilePath)
	if err != nil {
		log.Fatalf("Error creating file: %v", err)
	}
	defer file.Close()

	// Execute the template with the data and write it to the file
	err = tmpl.Execute(file, ki)
	if err != nil {
		log.Fatalf("Error executing template: %v", err)
	}

	return nil
}
