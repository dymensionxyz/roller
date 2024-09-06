package rollapp

import (
	"time"

	tmtypes "github.com/tendermint/tendermint/types"
)

type ShowRollappResponse struct {
	Rollapp Rollapp `protobuf:"bytes,1,opt,name=rollapp,proto3" json:"rollapp"`
	Summary Summary `protobuf:"bytes,6,opt,name=summary,proto3" json:"summary"`
	// apps is the list of (lazy-loaded) apps in the rollapp
}

type Summary struct {
	RollappId                 string          `json:"rollappId,omitempty"`
	LatestStateIndex          *StateInfoIndex `json:"latestStateIndex,omitempty"`
	LatestFinalizedStateIndex *StateInfoIndex `json:"latestFinalizedStateIndex,omitempty"`
	LatestHeight              string          `json:"latestHeight,omitempty"`
	LatestFinalizedHeight     string          `json:"latestFinalizedHeight,omitempty"`
}

type Rollapp struct {
	RollappId             string              `json:"rollapp_id,omitempty"`
	Owner                 string              `json:"owner,omitempty"`
	GenesisState          RollappGenesisState `json:"genesis_state"`
	ChannelId             string              `json:"channel_id,omitempty"`
	Frozen                bool                `json:"frozen,omitempty"`
	RegisteredDenoms      []string            `json:"registeredDenoms,omitempty"`
	Metadata              *RollappMetadata    `json:"metadata,omitempty"`
	GenesisInfo           GenesisInfo         `json:"genesis_info"`
	InitialSequencer      string              `json:"initial_sequencer,omitempty"`
	VmType                string              `json:"vm_type,omitempty"`
	Launched              bool                `json:"launched,omitempty"`
	LivenessEventHeight   string              `json:"liveness_event_height,omitempty"`
	LastStateUpdateHeight string              `json:"last_state_update_height,omitempty"`
}

type GenesisInfo struct {
	GenesisChecksum string         `json:"genesis_checksum,omitempty"`
	Bech32Prefix    string         `json:"bech32_prefix,omitempty"`
	NativeDenom     *DenomMetadata `json:"native_denom,omitempty"`
	InitialSupply   string         `json:"initial_supply"`
	Sealed          bool           `json:"sealed,omitempty"           protobuf:"varint,5,opt,name=sealed,proto3"`
}

type RollappMetadata struct {
	Website     string         `json:"website,omitempty"`
	Description string         `json:"description,omitempty"`
	LogoUrl     string         `json:"logo_url,omitempty"`
	Telegram    string         `json:"telegram,omitempty"`
	X           string         `json:"x,omitempty"`
	GenesisUrl  string         `json:"genesis_url,omitempty"`
	DisplayName string         `json:"display_name,omitempty"`
	Tagline     string         `json:"tagline,omitempty"`
	ExplorerUrl string         `json:"explorer_url,omitempty"`
	FeeDenom    *DenomMetadata `json:"fee_denom,omitempty"`
}

type DenomMetadata struct {
	Display  string `json:"display,omitempty"`
	Base     string `json:"base,omitempty"`
	Exponent uint32 `json:"exponent,omitempty"`
}

type RollappGenesisState struct {
	TransfersEnabled bool `json:"transfers_enabled,omitempty"`
}

type GenesisState struct {
	TransfersEnabled bool `json:"transfers_enabled"`
}

type Metadata struct {
	Website          string `json:"website"`
	Description      string `json:"description"`
	LogoUrl          string `json:"logo_url"`
	TokenLogoDataUri string `json:"token_logo_data_uri"`
	Telegram         string `json:"telegram"`
	X                string `json:"x"`
	GenesisUrl       string `json:"genesis_url"`
	TokenSymbol      string `json:"token_symbol"`
}

type StateInfoIndex struct {
	RollappId string `json:"rollappId,omitempty"`
	Index     string `json:"index,omitempty"`
}

type BlockInformation struct {
	BlockId tmtypes.BlockID `json:"block_id"`
	Block   Block           `json:"block"`
}

type Block struct {
	Header       `json:"header"`
	tmtypes.Data `json:"data"`
	Evidence     tmtypes.EvidenceData `json:"evidence"`
}

type Header struct {
	Version Consensus `json:"version"`
	ChainID string    `json:"chain_id"`
	Height  string    `json:"height"`
	Time    time.Time `json:"time"`
}

type Consensus struct {
	Block string `json:"block,omitempty"`
	App   string `json:"app,omitempty"`
}
