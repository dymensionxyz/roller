package rollapp

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
}

type StateInfoIndex struct {
	// rollappId is the rollapp that the sequencer belongs to and asking to update
	// it used to identify the what rollapp a StateInfo belongs
	// The rollappId follows the same standard as cosmos chain_id
	RollappId string `protobuf:"bytes,1,opt,name=rollappId,proto3" json:"rollappId,omitempty"`
	// index is a sequential increasing number, updating on each
	// state update used for indexing to a specific state info, the first index is 1
	Index uint64 `protobuf:"varint,2,opt,name=index,proto3"    json:"index,omitempty"`
}
