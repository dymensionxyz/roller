package claim

import (
	"fmt"
	"math/big"
	"os"
	"os/exec"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/keys"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim-rewards <private-key> <destination-address>",
		Short: "Send the DYM rewards associated with the given private key to the destination address",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			shouldProceed, err := pterm.DefaultInteractiveConfirm.WithDefaultText(
				fmt.Sprintf(
					"This command will transfer all Rollapp rewards on the mainnet to %s. Please note that once"+
						" initiated, this action cannot be undone. Do you wish to proceed",
					args[1],
				),
			).Show()
			if err != nil {
				return err
			}

			if !shouldProceed {
				return nil
			}
			tempDir, err := os.MkdirTemp(os.TempDir(), "hub_sequencer")
			if err != nil {
				return err
			}

			importKeyCmd := exec.Command(
				consts.Executables.Simd, "keys", "import-hex",
				consts.KeysIds.HubSequencer, args[0], "--home", tempDir,
			)
			_, err = bash.ExecCommandWithStdout(importKeyCmd)
			if err != nil {
				return err
			}

			sequencerAddr, err := keys.GetAddressBinary(
				keys.KeyConfig{
					ID:  consts.KeysIds.HubSequencer,
					Dir: tempDir,
				}, tempDir,
			)
			if err != nil {
				return err
			}

			mainnetHub := consts.Hubs[consts.MainnetHubName]
			sequencerBalance, err := keys.QueryBalance(
				keys.ChainQueryConfig{
					Binary: consts.Executables.Dymension,
					Denom:  consts.Denoms.Hub,
					RPC:    mainnetHub.RPC_URL,
				}, sequencerAddr,
			)
			if err != nil {
				return err
			}

			// Calculated by sending a tx on Froopyland and see how much fees were paid
			txGasPrice := big.NewInt(50000)
			totalBalanceMinusFees := new(big.Int).Sub(sequencerBalance.Amount, txGasPrice)
			if totalBalanceMinusFees.Cmp(big.NewInt(0)) != 1 {
				return fmt.Errorf(
					"no rewards to claim for the address associated with the given private key: %s"+
						"please try to import the private key to keplr and claim the rewards from there",
					sequencerAddr,
				)
			}
			rewardsAmountStr := totalBalanceMinusFees.String() + consts.Denoms.Hub

			sendAllFundsCmd := exec.Command(
				consts.Executables.Dymension,
				"tx",
				"bank",
				"send",
				consts.KeysIds.HubSequencer,
				args[1],
				rewardsAmountStr,
				"--node",
				mainnetHub.RPC_URL,
				"--chain-id",
				mainnetHub.ID,
				"--fees",
				txGasPrice.String()+consts.Denoms.Hub,
				"-b",
				"block",
				"--yes",
				"--home",
				tempDir,
			)
			_, err = bash.ExecCommandWithStdout(sendAllFundsCmd)
			if err != nil {
				return err
			}

			fmt.Printf("💈 Successfully claimed %s to %s!\n", rewardsAmountStr, args[1])
			return nil
		},
	}
	return cmd
}
