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
	// The unique identifier of the rollapp chain.
	// The rollappId follows the same standard as cosmos chain_id.
	RollappId string `protobuf:"bytes,1,opt,name=rollappId,proto3"                 json:"rollappId,omitempty"`
	// Defines the index of the last rollapp UpdateState.
	LatestStateIndex *StateInfoIndex `protobuf:"bytes,2,opt,name=latestStateIndex,proto3"          json:"latestStateIndex,omitempty"`
	// Defines the index of the last rollapp UpdateState that was finalized.
	LatestFinalizedStateIndex *StateInfoIndex `protobuf:"bytes,3,opt,name=latestFinalizedStateIndex,proto3" json:"latestFinalizedStateIndex,omitempty"`
	LatestHeight              uint64          `protobuf:"varint,4,opt,name=latestHeight,proto3"             json:"latestHeight,omitempty"`
	LatestFinalizedHeight     uint64          `protobuf:"varint,5,opt,name=latestFinalizedHeight,proto3"    json:"latestFinalizedHeight,omitempty"`
}

type Rollapp struct {
	RollappId             string              `json:"rollapp_id,omitempty"`
	Owner                 string              `json:"owner,omitempty"`
	GenesisState          RollappGenesisState `json:"genesis_state"`
	ChannelId             string              `json:"channel_id,omitempty"`
	Frozen                bool                `json:"frozen,omitempty"`
	RegisteredDenoms      []string            `json:"registeredDenoms,omitempty"`
	Bech32Prefix          string              `json:"bech32_prefix,omitempty"`
	GenesisChecksum       string              `json:"genesis_checksum,omitempty"`
	Metadata              *RollappMetadata    `json:"metadata,omitempty"`
	InitialSequencer      string              `json:"initial_sequencer,omitempty"`
	VmType                string              `json:"vm_type,omitempty"`
	Sealed                bool                `json:"sealed,omitempty"`
	LivenessEventHeight   string              `json:"liveness_event_height,omitempty"`
	LastStateUpdateHeight string              `json:"last_state_update_height,omitempty"`
}

type RollappMetadata struct {
	Website         string `json:"website,omitempty"`
	Description     string `json:"description,omitempty"`
	LogoUrl         string `json:"logo_url,omitempty"`
	Telegram        string `json:"telegram,omitempty"`
	X               string `json:"x,omitempty"`
	GenesisUrl      string `json:"genesis_url,omitempty"`
	DisplayName     string `json:"display_name,omitempty"`
	Tagline         string `json:"tagline,omitempty"`
	TokenSymbol     string `json:"token_symbol,omitempty"`
	ExplorerUrl     string `json:"explorer_url,omitempty"`
	FeeBaseDenom    string `json:"fee_base_denom,omitempty"`
	NativeBaseDenom string `json:"native_base_denom,omitempty"`
}

type RollappGenesisState struct {
	TransfersEnabled bool `protobuf:"varint,2,opt,name=transfers_enabled,json=transfersEnabled,proto3" json:"transfers_enabled,omitempty"`
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
	// rollappId is the rollapp that the sequencer belongs to and asking to update
	// it used to identify the what rollapp a StateInfo belongs
	// The rollappId follows the same standard as cosmos chain_id
	RollappId string `protobuf:"bytes,1,opt,name=rollappId,proto3" json:"rollappId,omitempty"`
	// index is a sequential increasing number, updating on each
	// state update used for indexing to a specific state info, the first index is 1
	Index string `protobuf:"varint,2,opt,name=index,proto3"    json:"index,omitempty"`
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
	// basic block info
	Version Consensus `json:"version"`
	ChainID string    `json:"chain_id"`
	Height  string    `json:"height"`
	Time    time.Time `json:"time"`
}

type Consensus struct {
	Block string `protobuf:"varint,1,opt,name=block,proto3" json:"block,omitempty"`
	App   string `protobuf:"varint,2,opt,name=app,proto3"   json:"app,omitempty"`
}
