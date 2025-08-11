package sequencer

import (
	"errors"
	"fmt"
	"strings"

	cosmossdkmath "cosmossdk.io/math"
	cosmossdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/denom"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/roller"
)

func CheckBalance(rollappConfig roller.RollappConfig) error {
	pterm.Info.Println("getting the existing sequencer address ")
	hubSeqKC := keys.KeyConfig{
		Dir:            consts.ConfigDirName.HubKeys,
		ID:             consts.KeysIds.HubSequencer,
		ChainBinary:    consts.Executables.Dymension,
		Type:           consts.SDK_ROLLAPP,
		KeyringBackend: rollappConfig.KeyringBackend,
	}
	seqAddrInfo, err := hubSeqKC.Info(rollappConfig.Home)
	if err != nil {
		return err
	}
	seqAddrInfo.Address = strings.TrimSpace(seqAddrInfo.Address)

	pterm.Info.Println("getting the existing sequencer address balance")
	balance, err := keys.QueryBalance(
		keys.ChainQueryConfig{
			Denom:  consts.Denoms.Hub,
			RPC:    rollappConfig.HubData.RpcUrl,
			Binary: consts.Executables.Dymension,
		}, seqAddrInfo.Address,
	)
	if err != nil {
		pterm.Error.Println("failed to get address balance: ", err)
		return err
	}

	opsAmnt, _ := cosmossdkmath.NewIntFromString(consts.MinOperationalAmount)

	desiredBond := cosmossdktypes.NewCoin(
		consts.Denoms.Hub,
		opsAmnt,
	)

	necessaryBalance := desiredBond.Amount

	necessaryBalance = necessaryBalance.Add(
		cosmossdkmath.NewInt(consts.DefaultTxFee),
	)

	blnc, _ := denom.BaseDenomToDenom(*balance, 18)
	oneDym, _ := cosmossdkmath.NewIntFromString("1000000000000000000")

	nb := cosmossdktypes.Coin{
		Denom:  consts.Denoms.Hub,
		Amount: necessaryBalance.Add(oneDym),
	}
	necBlnc, _ := denom.BaseDenomToDenom(nb, 18)

	pterm.Info.Printf(
		"current balance: %s (%s)\nnecessary balance: %s (%s)\n",
		balance.String(),
		blnc.String(),
		fmt.Sprintf("%s%s", necessaryBalance.String(), consts.Denoms.Hub),
		necBlnc.String(),
	)

	isAddrFunded := balance.Amount.GTE(necessaryBalance)
	if !isAddrFunded {
		pterm.DefaultSection.WithIndentCharacter("🔔").
			Println("Please fund the addresses below to register and run the sequencer.")
		seqAddrInfo.Print(keys.WithName())
		proceed, _ := pterm.DefaultInteractiveConfirm.WithDefaultValue(false).
			WithDefaultText(
				"press 'y' when the wallets are funded",
			).Show()

		if !proceed {
			return errors.New("cancelled by user")
		}
	}
	return nil
}
