package relayer

import (
	"slices"
	"time"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/roller"
	sequencerutils "github.com/dymensionxyz/roller/utils/sequencer"
)

func (r *Relayer) HandleWhitelisting(
	addr string,
	rollappChainData *roller.RollappConfig,
) error {
	kb := rollappChainData.KeyringBackend

	seqAddr, err := sequencerutils.GetSequencerAccountAddress(*rollappChainData)
	if err != nil {
		return err
	}

	isRlyKeyWhitelisted, err := isRelayerRollappKeyWhitelisted(
		seqAddr,
		addr,
		r.Hub,
	)
	if err != nil {
		return err
	}

	if !isRlyKeyWhitelisted {
		pterm.Warning.Printfln(
			"relayer key (%s) is not whitelisted, updating whitelisted relayers",
			addr,
		)

		err := sequencerutils.UpdateWhitelistedRelayers(
			r.RollerHome,
			addr,
			string(kb),
			r.Hub,
		)
		if err != nil {
			pterm.Error.Println("failed to update whitelisted relayers:", err)
			return err
		}
	}

	raOpAddr, err := sequencerutils.GetSequencerOperatorAddress(
		r.RollerHome,
		string(kb),
	)
	if err != nil {
		pterm.Error.Println("failed to get RollApp's operator address:", err)
		return err
	}

	wrSpinner, _ := pterm.DefaultSpinner.Start(
		"waiting for the whitelisted relayer to propagate to RollApp (this might take a while)",
	)
	for {
		wra, err := sequencerutils.GetWhitelistedRelayersOnRa(raOpAddr)
		if err != nil {
			pterm.Error.Println("failed to get whitelisted relayers for rollapp operator", raOpAddr, err)
			return err
		}

		if len(wra) == 0 &&
			slices.Contains(wra, addr) {
			wrSpinner.UpdateText(
				"waiting for the whitelisted relayer to propagate to RollApp...",
			)
			time.Sleep(time.Second * 5)
			continue
		} else {
			// nolint: errcheck
			wrSpinner.Success("relayer whitelisted and propagated to rollapp")
			break
		}
	}

	return nil
}

func isRelayerRollappKeyWhitelisted(seqAddr, relAddr string, hd consts.HubData) (bool, error) {
	relayers, err := sequencerutils.GetWhitelistedRelayersOnHub(seqAddr, hd)
	if err != nil {
		return false, err
	}

	return slices.Contains(relayers, relAddr), nil
}
