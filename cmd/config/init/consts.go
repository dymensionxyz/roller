package initconfig

var FlagNames = struct {
	TokenSupply   string
	RollappBinary string
	HubID         string
	Decimals      string
	Interactive   string
	DAType        string
	VMType        string
	NoOutput      string
}{
	TokenSupply:   "token-supply",
	RollappBinary: "rollapp-binary",
	HubID:         "hub",
	Interactive:   "interactive",
	Decimals:      "decimals",
	DAType:        "da",
	VMType:        "vm-type",
	NoOutput:      "no-output",
}

const (
	StagingHubName    = "devnet"
	FroopylandHubName = "froopyland"
	LocalHubName      = "local"
	LocalHubID        = "dymension_100-1"
)
