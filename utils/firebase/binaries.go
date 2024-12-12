package firebase

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

// BinaryVersionInfo represents the version information for DRS binaries
type BinaryVersionInfo struct {
	EibcClient string `firestore:"eibc-client"`
	Relayer    string `firestore:"relayer"`
}

// GetDependencyVersions
// Fetch dependency binary versions using the nested collection path
// Path format: tool-versions/latest/<doc>
func GetDependencyVersions() (*BinaryVersionInfo, error) {
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

	var bvi BinaryVersionInfo
	if err := docSnapshot.DataTo(&bvi); err != nil {
		return nil, fmt.Errorf("failed to parse DRS version info: %v", err)
	}

	return &bvi, nil
}
