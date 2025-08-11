package denom

import (
	cosmossdkmath "cosmossdk.io/math"
	cosmossdktypes "github.com/cosmos/cosmos-sdk/types"
)

func BaseDenomToDenom(
	coin cosmossdktypes.Coin,
	exponent int,
) (cosmossdktypes.Coin, error) {
	exp := cosmossdkmath.NewIntWithDecimal(1, exponent)

	coin.Amount = coin.Amount.Quo(exp)
	coin.Denom = coin.Denom[1:]

	return coin, nil
}

func DenomToBaseDenom(
	coin cosmossdktypes.Coin,
	exponent int,
) (cosmossdktypes.Coin, error) {
	exp := cosmossdkmath.NewIntWithDecimal(1, exponent)

	coin.Amount = coin.Amount.Mul(exp)

	return coin, nil
}
