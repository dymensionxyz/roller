package oracle

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func (o *OracleConfig) DownloadContractCode() error {
	contractURL := "https://storage.googleapis.com/dymension-roller/centralized_oracle.wasm"
	contractPath := filepath.Join(o.ConfigDirPath, "centralized_oracle.wasm")

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(o.ConfigDirPath, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	// Create the file
	out, err := os.Create(contractPath)
	if err != nil {
		return fmt.Errorf("failed to create contract file: %v", err)
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(contractURL)
	if err != nil {
		return fmt.Errorf("failed to download contract: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download contract, status: %s", resp.Status)
	}

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save contract: %v", err)
	}

	return nil
}
