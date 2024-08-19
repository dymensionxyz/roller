package sequencer

import (
	"time"

	cosmossdkmath "cosmossdk.io/math"
	cosmossdktypes "github.com/cosmos/cosmos-sdk/types"
	dymensionseqtypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/gogo/protobuf/types"
)

type Sequencers struct {
	Sequencers []Info `json:"sequencers,omitempty"`
}

type ShowSequencerResponse struct {
	Sequencer Info `json:"sequencers,omitempty"`
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
}

type Metadata struct {
	// moniker defines a human-readable name for the sequencer.
	Moniker string `protobuf:"bytes,1,opt,name=moniker,proto3"                                                                    json:"moniker,omitempty"`
	// details define other optional details.
	Details string `protobuf:"bytes,5,opt,name=details,proto3"                                                                    json:"details,omitempty"`
	// bootstrap nodes list
	P2PSeeds []string `protobuf:"bytes,6,rep,name=p2p_seeds,json=p2pSeeds,proto3"                                                    json:"p2p_seeds,omitempty"`
	// RPCs list
	Rpcs []string `protobuf:"bytes,7,rep,name=rpcs,proto3"                                                                       json:"rpcs,omitempty"`
	// evm RPCs list
	EvmRpcs []string `protobuf:"bytes,8,rep,name=evm_rpcs,json=evmRpcs,proto3"                                                      json:"evm_rpcs,omitempty"`
	// REST API URLs
	RestApiUrls []string `protobuf:"bytes,9,rep,name=rest_api_urls,json=restApiUrls,proto3"                                             json:"rest_api_urls,omitempty"`
	// block explorer URL
	ExplorerUrl string `protobuf:"bytes,10,opt,name=explorer_url,json=explorerUrl,proto3"                                             json:"explorer_url,omitempty"`
	// genesis URLs
	GenesisUrls []string `protobuf:"bytes,11,rep,name=genesis_urls,json=genesisUrls,proto3"                                             json:"genesis_urls,omitempty"`
	// contact details
	// nolint:govet,staticcheck
	ContactDetails *dymensionseqtypes.ContactDetails `protobuf:"bytes,12,opt,name=contact_details,
json=contactDetails,
proto3"                                       json:"contact_details,omitempty"`
	// json dump the sequencer can add (limited by size)
	ExtraData []byte `protobuf:"bytes,13,opt,name=extra_data,json=extraData,proto3"                                                 json:"extra_data,omitempty"`
	// snapshots of the sequencer
	Snapshots []*SnapshotInfo `protobuf:"bytes,14,rep,name=snapshots,proto3"                                                                 json:"snapshots,omitempty"`
	// gas_price defines the value for each gas unit
	// nolint:govet,staticcheck
	GasPrice *cosmossdkmath.Int `protobuf:"bytes,15,opt,name=gas_price,json=gasPrice,proto3,
customtype=github.com/cosmos/cosmos-sdk/types.Int" json:"gas_price,omitempty"`
}

type SnapshotInfo struct {
	// the snapshot url
	SnapshotUrl string `protobuf:"bytes,1,opt,name=snapshot_url,json=snapshotUrl,proto3" json:"snapshot_url,omitempty"`
	// The snapshot height
	Height string `protobuf:"varint,2,opt,name=height,proto3"                       json:"height,omitempty"`
	// sha-256 checksum value for the snapshot file
	Checksum string `protobuf:"bytes,3,opt,name=checksum,proto3"                      json:"checksum,omitempty"`
}
