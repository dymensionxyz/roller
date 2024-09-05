package blockexplorer

import (
	"fmt"
	"os"
	"path/filepath"
)

// GenerateChainsYAML generates the YAML content with the given chain_id
// this configuration is used by the block-explorer to index the locally running
// chain
func GenerateChainsYAML(chainID string) string {
	template := `local:
  chain_id: %s
  be_json_rpc_urls: [ "http://host.docker.internal:11100" ]
  # disable: true
`
	return fmt.Sprintf(template, chainID)
}

func WriteChainsYAML(filePath, content string) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o700)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}
