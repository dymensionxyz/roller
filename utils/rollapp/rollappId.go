package rollapp

import (
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	errorsmod "cosmossdk.io/errors"
	dymratypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/pterm/pterm"
)

const (
	// MaxChainIDLen is a maximum length of the chain ID.
	MaxChainIDLen = 50
)

var (
	regexChainID         = `[a-z]{1,}`
	regexEIP155Separator = `_{1}`
	regexEIP155          = `[1-9][0-9]*`
	regexEpochSeparator  = `-{1}`
	regexEpoch           = `[1-9][0-9]*`
	ethermintChainID     = regexp.MustCompile(
		fmt.Sprintf(
			`^(%s)%s(%s)%s(%s)$`,
			regexChainID,
			regexEIP155Separator,
			regexEIP155,
			regexEpochSeparator,
			regexEpoch,
		),
	)
)

type ChainID struct {
	chainID  string
	name     string
	eip155ID *big.Int
	revision uint64
}

// from Dymension source code `x/rollapp/types/chain_id.go` @20240911
func ValidateChainID(id string) (ChainID, error) {
	spinner, _ := pterm.DefaultSpinner.WithRemoveWhenDone().
		Start(fmt.Sprintf("validating rollapp id '%s'\n", id))
	chainID := strings.TrimSpace(id)

	if chainID == "" {
		return ChainID{}, errorsmod.Wrapf(dymratypes.ErrInvalidRollappID, "empty")
	}

	if len(chainID) > MaxChainIDLen {
		// nolint: errcheck,gosec
		spinner.Stop()
		return ChainID{}, errorsmod.Wrapf(
			dymratypes.ErrInvalidRollappID,
			"exceeds %d chars: %s: len: %d",
			MaxChainIDLen,
			chainID,
			len(chainID),
		)
	}

	matches := ethermintChainID.FindStringSubmatch(chainID)

	if matches == nil || len(matches) != 4 || matches[1] == "" {
		return ChainID{}, dymratypes.ErrInvalidRollappID
	}
	// verify that the chain-id entered is a base 10 integer
	chainIDInt, ok := new(big.Int).SetString(matches[2], 10)
	if !ok {
		// nolint: errcheck,gosec
		spinner.Stop()
		return ChainID{}, errorsmod.Wrapf(
			dymratypes.ErrInvalidRollappID,
			"EIP155 part %s must be base-10 integer format",
			matches[2],
		)
	}

	revision, err := strconv.ParseUint(matches[3], 0, 64)
	if err != nil {
		// nolint: errcheck,gosec
		spinner.Stop()
		return ChainID{}, errorsmod.Wrapf(
			dymratypes.ErrInvalidRollappID,
			"parse revision number: error: %v",
			err,
		)
	}

	spinner.Success(fmt.Sprintf("'%s' is a valid RollApp ID", id))
	return ChainID{
		chainID:  chainID,
		eip155ID: chainIDInt,
		revision: revision,
		name:     matches[1],
	}, nil
}
