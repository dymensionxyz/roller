package structs

import (
	"encoding/json"
	"fmt"
	"os"

	cosmossdkmath "cosmossdk.io/math"
	sequencerutils "github.com/dymensionxyz/roller/utils/sequencer"
)

func InitializeMetadata(m sequencerutils.Metadata) {
	if m.Moniker == "" {
		m.Moniker = ""
	}
	if m.Details == "" {
		m.Details = ""
	}
	if m.P2PSeeds == nil {
		m.P2PSeeds = []string{}
	}
	if m.Rpcs == nil {
		m.Rpcs = []string{}
	}
	if m.EvmRpcs == nil {
		m.EvmRpcs = []string{}
	}
	if m.RestApiUrls == nil {
		m.RestApiUrls = []string{}
	}
	if m.ExplorerUrl == "" {
		m.ExplorerUrl = ""
	}
	if m.GenesisUrls == nil {
		m.GenesisUrls = []string{}
	}
	if m.ContactDetails == nil {
		m.ContactDetails = &sequencerutils.ContactDetails{}
	}
	if m.ExtraData == nil {
		m.ExtraData = []byte{}
	}
	if m.Snapshots == nil {
		m.Snapshots = []*sequencerutils.SnapshotInfo{}
	}
	if m.GasPrice == nil {
		zero := cosmossdkmath.NewInt(0)
		m.GasPrice = &zero
	}
}

func ExportStructToFile(data sequencerutils.Metadata, filename string) error {
	// Initialize the struct with default values
	InitializeMetadata(data)

	// Marshal the struct to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling to JSON: %v", err)
	}

	// Write to file
	err = os.WriteFile(filename, jsonData, 0o644)
	if err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}

	return nil
}
