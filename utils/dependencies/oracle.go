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
	"github.com/dymensionxyz/roller/utils/archives"
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
	PriceEvmOracle  string `firestore:"price-evm"`
	PriceWasmOracle string `firestore:"price-wasm"`
	RngEvmOracle    string `firestore:"rng-evm"`
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
		Doc("oracle-client-binaries")

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

func InstallBinary(ctx context.Context, config BinaryInstallConfig, oracleType string) error {
	// Determine oracle type based on rollapp type
	var oraclePlatform, oraclePath string
	switch {
	case config.RollappType == consts.WASM_ROLLAPP && oracleType == "price":
		oraclePlatform = "cosmos-oracle"
		oraclePath = "price/cosmos"
	case config.RollappType == consts.WASM_ROLLAPP && oracleType == "rng":
		oraclePlatform = "cosmos-oracle"
		oraclePath = "rng/cosmos"
	case config.RollappType == consts.EVM_ROLLAPP && oracleType == "price":
		oraclePlatform = "evm-oracle"
		oraclePath = "price/evm"
	case config.RollappType == consts.EVM_ROLLAPP && oracleType == "rng":
		oraclePlatform = "evm-oracle"
		oraclePath = "rng/evm"
	}

	// Build the artifact path
	arch := runtime.GOARCH
	goos := runtime.GOOS
	version := fmt.Sprintf("%s-%s", config.Version, arch)

	binaryName := fmt.Sprintf("%s-%s-%s", oraclePlatform, goos, arch)

	fullPath := fmt.Sprintf("%s/%s/%s/%s",
		oraclePlatform, // package name (cosmos-oracle)
		version,        // version (1.0.0-amd64)
		oraclePath,     // path (price/cosmos or rng/cosmos)
		binaryName,     // binary name (evm-oracle-linux-amd64)
	)

	downloadURL := fmt.Sprintf(
		"https://europe-central2-generic.pkg.dev/dymension-ops/oracle-clients/%s",
		fullPath,
	)

	fmt.Println("Downloading", downloadURL)
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

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "oracle-binary-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath) // Clean up temp file in case of failure

	// Copy downloaded content to temporary file
	if _, err := io.Copy(tmpFile, httpResp.Body); err != nil {
		return fmt.Errorf("failed to write binary to temporary file: %w", err)
	}
	tmpFile.Close()

	// Move the binary into place and make it executable
	if err := archives.MoveBinaryIntoPlaceAndMakeExecutable(tmpPath, config.InstallDir); err != nil {
		return fmt.Errorf("failed to move binary into place: %w", err)
	}

	pterm.Success.Printf("%s binary installed successfully at %s\n",
		strings.Title(oracleType), config.InstallDir)
	return nil
}
