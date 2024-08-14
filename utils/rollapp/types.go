package rollapp

import (
	"time"

	cmtbytes "github.com/tendermint/tendermint/libs/bytes"
	tmtypes "github.com/tendermint/tendermint/types"
)

type ShowRollappResponse struct {
	Rollapp                   Rollapp         `json:"rollapp"`
	LatestStateIndex          *StateInfoIndex `json:"latestStateIndex"`
	LatestFinalizedStateIndex *StateInfoIndex `json:"latestFinalizedStateIndex"`
	LatestHeight              string          `json:"latestHeight"`
	LatestFinalizedHeight     string          `json:"latestFinalizedHeight"`
}

type Rollapp struct {
	RollappId             string       `json:"rollapp_id"`
	Owner                 string       `json:"owner"`
	GenesisState          GenesisState `json:"genesis_state"`
	ChannelId             string       `json:"channel_id"`
	Frozen                bool         `json:"frozen"`
	RegisteredDenoms      []string     `json:"registeredDenoms"`
	Bech32Prefix          string       `json:"bech32_prefix"`
	GenesisChecksum       string       `json:"genesis_checksum"`
	Metadata              Metadata     `json:"metadata"`
	InitialSequencer      string       `json:"initial_sequencer"`
	VmType                string       `json:"vm_type"`
	Sealed                bool         `json:"sealed"`
	LivenessEventHeight   string       `json:"liveness_event_height"`
	LastStateUpdateHeight string       `json:"last_state_update_height"`
}

type GenesisState struct {
	TransfersEnabled bool `json:"transfers_enabled"`
}

type Metadata struct {
	Website          string `json:"website"`
	Description      string `json:"description"`
	LogoDataUri      string `json:"logo_data_uri"`
	TokenLogoDataUri string `json:"token_logo_data_uri"`
	Telegram         string `json:"telegram"`
	X                string `json:"x"`
	GenesisUrl       string `json:"genesis_url"`
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
	LastCommit   *tmtypes.Commit      `json:"last_commit"`
}

type Header struct {
	// basic block info
	Version Consensus `json:"version"`
	ChainID string    `json:"chain_id"`
	Height  int64     `json:"height"`
	Time    time.Time `json:"time"`

	// prev block info
	LastBlockID tmtypes.BlockID `json:"last_block_id"`

	// hashes of block data
	LastCommitHash cmtbytes.HexBytes `json:"last_commit_hash"` // commit from validators from the last block
	DataHash       cmtbytes.HexBytes `json:"data_hash"`        // transactions

	// hashes from the app output from the prev block
	ValidatorsHash     cmtbytes.HexBytes `json:"validators_hash"`      // validators for the current block
	NextValidatorsHash cmtbytes.HexBytes `json:"next_validators_hash"` // validators for the next block
	ConsensusHash      cmtbytes.HexBytes `json:"consensus_hash"`       // consensus params for current block
	AppHash            cmtbytes.HexBytes `json:"app_hash"`             // state after txs from the previous block
	// root hash of all results from the txs from the previous block
	// see `deterministicResponseDeliverTx` to understand which parts of a tx is hashed into here
	LastResultsHash cmtbytes.HexBytes `json:"last_results_hash"`

	// consensus info
	EvidenceHash    cmtbytes.HexBytes `json:"evidence_hash"`    // evidence included in the block
	ProposerAddress tmtypes.Address   `json:"proposer_address"` // original proposer of the block
}

type Consensus struct {
	Block string `protobuf:"varint,1,opt,name=block,proto3" json:"block,omitempty"`
	App   uint64 `protobuf:"varint,2,opt,name=app,proto3"   json:"app,omitempty"`
}
