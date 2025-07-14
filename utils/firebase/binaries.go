package firebase

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

// BinaryVersionInfo represents the version information for DRS binaries
type BinaryVersionInfo struct {
	EibcClient   string `firestore:"eibc-client"`
	Relayer      string `firestore:"relayer"`
	Dymd         string `firestore:"dymd"`
	CelestiaNode string `firestore:"celestia-node"`
	CelestiaApp  string `firestore:"celestia-app"`
}

// GetDependencyVersions
// Fetch dependency binary versions using the nested collection path
// Path format: <env>/tool-versions/field
func GetDependencyVersions(env string) (*BinaryVersionInfo, error) {
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

	var collectionName string

	if env == "mainnet" {
		collectionName = "mainnet"
	} else {
		collectionName = "testnets"
	}

	// Fetch DRS version information using the nested collection path
	// Path format: versions/{version}/revisions/{revision}
	drsDoc := client.Collection(collectionName).
		Doc("tool-versions")

	docSnapshot, err := drsDoc.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DRS version info: %v", err)
	}

	var bvi BinaryVersionInfo
	if err := docSnapshot.DataTo(&bvi); err != nil {
		return nil, fmt.Errorf("failed to parse DRS version info: %v", err)
	}

	return &bvi, nil
}
