package claim

import (
	"fmt"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/spf13/cobra"
	"math/big"
	"os"
	"os/exec"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim-rewards <private-key> <destination address>",
		Short: "Send the DYM rewards associated with the given private key to the destination address",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Replace with mainnet hub RPC
			mainnetRPC := consts.Hubs[consts.FroopylandHubName].RPC_URL
			tempDir, err := os.MkdirTemp(os.TempDir(), "hub_sequencer")
			if err != nil {
				return err
			}
			importKeyCmd := exec.Command(consts.Executables.Simd, "keys", "import-hex",
				consts.KeysIds.HubSequencer, args[0], "--home", tempDir)
			_, err = utils.ExecBashCommandWithStdout(importKeyCmd)
			if err != nil {
				return err
			}
			sequencerAddr, err := utils.GetAddressBinary(utils.KeyConfig{
				ID:  consts.KeysIds.HubSequencer,
				Dir: tempDir,
			}, consts.Executables.Dymension)
			if err != nil {
				return err
			}
			sequencerBalance, err := utils.QueryBalance(utils.ChainQueryConfig{
				Binary: consts.Executables.Dymension,
				Denom:  consts.Denoms.Hub,
				RPC:    mainnetRPC,
			}, sequencerAddr)
			if err != nil {
				return err
			}
			// Calculated by sending a tx on Froopyland and see how much fees were paid
			txGasPrice := big.NewInt(50000)
			totalBalanceMinusFees := new(big.Int).Sub(sequencerBalance.Amount, txGasPrice)
			if totalBalanceMinusFees.Cmp(big.NewInt(0)) != 1 {
				return fmt.Errorf("no rewards to claim for the address associated with the given private key: %s"+
					"please try to import the private key to keplr and claim the rewards from there",
					sequencerAddr)
			}
			return nil
		},
	}
	return cmd
}
