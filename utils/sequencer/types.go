package sequencer

import (
	"time"

	cosmossdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/types"
)

type Sequencers struct {
	Sequencers []Info `json:"sequencers,omitempty"`
}

type ShowSequencerResponse struct {
	Sequencer Info `json:"sequencer,omitempty"`
}

type Info struct {
	// address is the bech32-encoded address of the sequencer account which is the account that the message was sent from.
	Address string `protobuf:"bytes,1,opt,name=address,proto3"                                                       json:"address,omitempty"`
	// pubkey is the public key of the sequencers' dymint client, as a Protobuf Any.
	DymintPubKey *types.Any `protobuf:"bytes,2,opt,name=dymintPubKey,proto3"                                                  json:"dymintPubKey,omitempty"`
	// rollappId defines the rollapp to which the sequencer belongs.
	RollappId string `protobuf:"bytes,3,opt,name=rollappId,proto3"                                                     json:"rollappId,omitempty"`
	// metadata defines the extra information for the sequencer.
	Metadata Metadata `protobuf:"bytes,4,opt,name=metadata,proto3"                                                      json:"metadata"`
	// jailed defined whether the sequencer has been jailed from bonded status or not.
	Jailed bool `protobuf:"varint,5,opt,name=jailed,proto3"                                                       json:"jailed,omitempty"`
	// proposer defines whether the sequencer is a proposer or not.
	Proposer bool `protobuf:"varint,6,opt,name=proposer,proto3"                                                     json:"proposer,omitempty"`
	// status is the sequencer status (bonded/unbonding/unbonded).
	Status string `protobuf:"varint,7,opt,name=status,proto3,enum=dymensionxyz.dymension.sequencer.OperatingStatus" json:"status,omitempty"`
	// tokens define the delegated tokens (incl. self-delegation).
	Tokens cosmossdktypes.Coins `protobuf:"bytes,8,rep,name=tokens,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins"  json:"tokens"`
	// unbonding_height defines, if unbonding, the height at which this sequencer has begun unbonding.
	UnbondingHeight string `protobuf:"varint,9,opt,name=unbonding_height,json=unbondingHeight,proto3"                        json:"unbonding_height,omitempty"`
	// unbond_time defines, if unbonding, the min time for the sequencer to complete unbonding.
	UnbondTime time.Time `protobuf:"bytes,10,opt,name=unbond_time,json=unbondTime,proto3,stdtime"                          json:"unbond_time"`
	// WhitelistedRelayers is an array of the whitelisted relayer addresses. Addresses are bech32-encoded strings.
	WhitelistedRelayers []string `protobuf:"bytes,13,rep,name=whitelisted_relayers,json=whitelistedRelayers,proto3"                json:"whitelisted_relayers,omitempty"`
	// opted in defines whether the sequencer can be selected as proposer
	OptedIn bool `protobuf:"varint,14,opt,name=opted_in,proto3"                                                    json:"opted_in,omitempty"`
}

type Metadata struct {
	// moniker defines a human-readable name for the sequencer.
	Moniker string `json:"moniker"`
	// details define other optional details.
	Details string `json:"details"`
	// bootstrap nodes list
	P2PSeeds []string `json:"p2p_seeds"`
	// RPCs list
	Rpcs []string `json:"rpcs"`
	// evm RPCs list
	EvmRpcs []string `json:"evm_rpcs"`
	// REST API URLs
	RestApiUrls []string `json:"rest_api_urls"`
	// block explorer URL
	ExplorerUrl string `json:"explorer_url"`
	// genesis URLs
	GenesisUrls []string `json:"genesis_urls"`
	// contact details
	// nolint:govet,staticcheck
	ContactDetails *ContactDetails `json:"contact_details"`
	// json dump the sequencer can add (limited by size)
	ExtraData []byte `json:"extra_data"`
	// snapshots of the sequencer
	Snapshots []*SnapshotInfo `json:"snapshots"`
	// gas_price defines the value for each gas unit
	// nolint:govet,staticcheck
	GasPrice string         `json:"gas_price"`
	FeeDenom *DenomMetadata `json:"fee_denom"`
}

type DenomMetadata struct {
	Display  string `json:"display"`
	Base     string `json:"base"`
	Exponent int    `json:"exponent"`
}

type ContactDetails struct {
	// website URL
	Website string `json:"website"`
	// telegram link
	Telegram string `json:"telegram"`
	// twitter link
	X string `json:"x"`
}

type SnapshotInfo struct {
	// the snapshot url
	SnapshotUrl string `protobuf:"bytes,1,opt,name=snapshot_url,json=snapshotUrl,proto3" json:"snapshot_url,omitempty"`
	// The snapshot height
	Height string `protobuf:"varint,2,opt,name=height,proto3"                       json:"height,omitempty"`
	// sha-256 checksum value for the snapshot file
	Checksum string `protobuf:"bytes,3,opt,name=checksum,proto3"                      json:"checksum,omitempty"`
}

type CheckExistingSequencerResponse struct {
	IsSequencerAlreadyRegistered bool
	IsSequencerKeyPresent        bool
	IsSequencerProposer          bool
}
