package dependencies

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	firebase "firebase.google.com/go"
	"github.com/pterm/pterm"
	"google.golang.org/api/option"

	"github.com/dymensionxyz/roller/cmd/consts"
)

const (
	projectID  = "dymension-ops"
	location   = "europe-central2"
	repository = "oracle-clients"
)

type BinaryInstallConfig struct {
	RollappType consts.VMType
	Version     string
	InstallDir  string
}

type OracleBinaryVersionInfo struct {
	EvmOracle  string `firestore:"evm-oracle"`
	WasmOracle string `firestore:"wasm-oracle"`
}

func GetOracleBinaryVersion(vmt consts.VMType) (*OracleBinaryVersionInfo, error) {
	ctx := context.Background()
	conf := &firebase.Config{ProjectID: "drs-metadata"}
	app, err := firebase.NewApp(ctx, conf, option.WithoutAuthentication())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize firebase app: %v", err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create firestore client: %v", err)
	}
	defer client.Close()

	// Fetch DRS version information using the nested collection path
	// Path format: versions/{version}/revisions/{revision}
	drsDoc := client.Collection("tool-versions").
		Doc("latest")

	docSnapshot, err := drsDoc.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DRS version info: %v", err)
	}

	var bvi OracleBinaryVersionInfo
	if err := docSnapshot.DataTo(&bvi); err != nil {
		return nil, fmt.Errorf("failed to parse DRS version info: %v", err)
	}

	return &bvi, nil
}

func InstallBinary(ctx context.Context, config BinaryInstallConfig) error {
	// Determine oracle type based on rollapp type
	oracleType := "evm-oracle"
	if config.RollappType == consts.WASM_ROLLAPP {
		oracleType = "cosmos-oracle"
	}

	// Build the artifact path
	arch := runtime.GOARCH
	goos := runtime.GOOS

	version := fmt.Sprintf("%s-%s", config.Version, arch)

	fullPath := fmt.Sprintf("%s:%s:%s-%s-%s",
		oracleType, // cosmos-oracle
		version,    // 1.0.0-amd64
		oracleType, // cosmos-oracle
		goos,       // linux
		arch,       // amd64
	)

	downloadURL := fmt.Sprintf(
		"https://artifactregistry.googleapis.com/download/v1/projects/%s/locations/%s/repositories/%s/files/%s:download?alt=media",
		projectID,
		location,
		repository,
		fullPath,
	)

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create download request: %w", err)
	}

	// Make the request
	httpClient := &http.Client{}
	httpResp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer httpResp.Body.Close()

	// Create installation directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(config.InstallDir), 0o755); err != nil {
		return fmt.Errorf("failed to create installation directory: %w", err)
	}

	// Save the file
	destFile, err := os.OpenFile(config.InstallDir, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, httpResp.Body); err != nil {
		return fmt.Errorf("failed to write binary: %w", err)
	}

	pterm.Success.Printf("%s binary installed successfully at %s\n",
		strings.Title(oracleType), config.InstallDir)
	return nil
}
