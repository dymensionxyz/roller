package firebase

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/pterm/pterm"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// DrsVersionInfo represents the structure of DRS version information in Firestore
type DrsVersionInfo struct {
	Commit string `firestore:"commit"`
}

// GetLatestDrsVersionCommit
// Fetch DRS version information using the nested collection path
// Path format: versions/{version}/revisions/{revision}
func GetLatestDrsVersionCommit(drsVersion string) (*DrsVersionInfo, error) {
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
	drsDoc := client.Collection("versions").
		Doc(drsVersion).
		Collection("revisions").
		OrderBy("timestamp", firestore.Desc).
		Limit(1).
		Documents(ctx)

	doc, err := drsDoc.Next()
	if err == iterator.Done {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	var drsInfo DrsVersionInfo
	if err := doc.DataTo(&drsInfo); err != nil {
		return nil, fmt.Errorf("failed to parse DRS version info: %v", err)
	}

	pterm.Info.Printf("Found DRS commit hash: %s\n", drsInfo.Commit)
	return &drsInfo, nil
}
