package jsonconfig

import (
	"os"

	"github.com/tidwall/sjson"

	"github.com/dymensionxyz/roller/utils/config"
)

// TODO(#130): fix to support epochs
func UpdateJSONParams(jsonFilePath string, params []config.PathValue) error {
	jsonFileContent, err := os.ReadFile(jsonFilePath)
	if err != nil {
		return err
	}
	jsonFileContentString := string(jsonFileContent)
	for _, param := range params {
		jsonFileContentString, err = sjson.Set(jsonFileContentString, param.Path, param.Value)
		if err != nil {
			return err
		}
	}

	// nolint:gofumpt
	err = os.WriteFile(jsonFilePath, []byte(jsonFileContentString), 0o644)
	if err != nil {
		return err
	}
	return nil
}
