package firebase

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

// DrsVersionInfo represents the structure of DRS version information in Firestore
type DrsVersionInfo struct {
	Commit     string `firestore:"commit"`
	EvmCommit  string `firestore:"evm-commit"`
	WasmCommit string `firestore:"wasm-commit"`
}

// GetLatestDrsVersionCommit
// Fetch DRS version information using the nested collection path
// Path format: versions/{version}/revisions/{revision}
// Path format: /testnets/versions/{version}/revisions/{revision}/version-info
// Path format: /mainnet/versions/{version}/revisions/{revision}/version-info
func GetLatestDrsVersionCommit(drsVersion string, env string) (*DrsVersionInfo, error) {
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

	var rootCollection string
	if env == "mainnet" {
		rootCollection = "mainnet"
	} else {
		rootCollection = "testnets"
	}

	// Get the path to revisions
	versionsDoc := client.Collection(rootCollection).Doc("versions")
	drsVersionColl := versionsDoc.Collection(drsVersion)

	// Find the highest revision number by trying incrementing numbers
	var latestDoc *firestore.DocumentSnapshot
	for i := 1; ; i++ {
		doc, err := drsVersionColl.Doc("revisions").
			Collection(fmt.Sprintf("%d", i)).
			Doc("version-info").
			Get(ctx)
		if err != nil {
			// If we get an error, we've gone past the last valid revision
			break
		}
		latestDoc = doc
	}

	if latestDoc == nil {
		return nil, fmt.Errorf("no version-info documents found for version %s", drsVersion)
	}

	var info DrsVersionInfo
	if err := latestDoc.DataTo(&info); err != nil {
		return nil, fmt.Errorf("failed to parse DRS version info: %v", err)
	}
	return &info, nil
}
