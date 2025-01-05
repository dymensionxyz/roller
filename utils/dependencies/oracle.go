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
	artifactregistry "google.golang.org/api/artifactregistry/v1"
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
	// Initialize client without credentials (assuming default application credentials)
	client, err := artifactregistry.NewService(
		ctx,
		option.WithoutAuthentication(),
	)
	if err != nil {
		return fmt.Errorf("failed to create artifact registry client: %w", err)
	}

	// Determine oracle type based on rollapp type
	oracleType := "evm-oracle"
	if config.RollappType == consts.WASM_ROLLAPP {
		oracleType = "cosmos-oracle"
	}

	// Build the artifact path
	arch := runtime.GOARCH
	goos := runtime.GOOS

	version := fmt.Sprintf("%s-%s", config.Version, arch)

	// Create parent path
	parent := fmt.Sprintf("projects/%s/locations/%s/repositories/%s",
		projectID, location, repository)

	// List files matching our criteria
	filesService := artifactregistry.NewProjectsLocationsRepositoriesFilesService(client)
	resp, err := filesService.List(parent).Filter(
		fmt.Sprintf("name.contains('%s')", version),
	).Do()
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	if len(resp.Files) == 0 {
		return fmt.Errorf("no binary found for %s oracle (version %s) on %s/%s",
			oracleType, version, goos, arch)
	}

	// Get download URL
	downloadPath := fmt.Sprintf("%s/%s", parent, resp.Files[0].Name)
	downloadResp, err := filesService.Download(downloadPath).Do()
	if err != nil {
		return fmt.Errorf("failed to get download URL: %w", err)
	}

	// Create installation directory
	if err := os.MkdirAll(filepath.Dir(config.InstallDir), 0o755); err != nil {
		return fmt.Errorf("failed to create installation directory: %w", err)
	}

	// Download the file
	pterm.Info.Printf("Downloading %s binary...\n", oracleType)
	downloadUrl := downloadResp.ServerResponse.Header.Get("X-Goog-Download-URL")
	if downloadUrl == "" {
		return fmt.Errorf("download URL not found in response headers")
	}
	// nolint:gosec
	httpResp, err := http.Get(downloadUrl)
	if err != nil {
		return fmt.Errorf("failed to download binary: %w", err)
	}
	defer httpResp.Body.Close()

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
