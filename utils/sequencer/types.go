package sequencer

import (
	"time"

	cosmossdktypes "github.com/cosmos/cosmos-sdk/types"
	dymensionseqtypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/gogo/protobuf/types"
)

type Info struct {
	// address is the bech32-encoded address of the sequencer account which is the account that the message was sent from.
	Address string `protobuf:"bytes,1,opt,name=address,proto3"                                                       json:"address,omitempty"`
	// pubkey is the public key of the sequencers' dymint client, as a Protobuf Any.
	DymintPubKey *types.Any `protobuf:"bytes,2,opt,name=dymintPubKey,proto3"                                                  json:"dymintPubKey,omitempty"`
	// rollappId defines the rollapp to which the sequencer belongs.
	RollappId string `protobuf:"bytes,3,opt,name=rollappId,proto3"                                                     json:"rollappId,omitempty"`
	// metadata defines the extra information for the sequencer.
	Metadata dymensionseqtypes.SequencerMetadata `protobuf:"bytes,4,opt,name=metadata,proto3"                                                      json:"metadata"`
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
