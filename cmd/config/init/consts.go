package initconfig

var FlagNames = struct {
	TokenSupply   string
	RollappBinary string
	HubID         string
	Decimals      string
	Interactive   string
	DAType        string
}{
	TokenSupply:   "token-supply",
	RollappBinary: "rollapp-binary",
	HubID:         "hub",
	Interactive:   "interactive",
	Decimals:      "decimals",
	DAType:        "da",
}

const (
	StagingHubName = "devnet"
	LocalHubName   = "local"
)
