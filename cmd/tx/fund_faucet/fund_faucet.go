package fund_faucet

import (
	"errors"
	"fmt"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/tx/tx_utils"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"github.com/dymensionxyz/roller/relayer"
	"github.com/spf13/cobra"
	"math/big"
	"os/exec"
	"path/filepath"
)

var flagNames = struct {
	tokenAmount string
	NoOutput    string
}{
	tokenAmount: "token-amount",
	NoOutput:    "no-output",
}

const faucetAddr = "dym1g8sf7w4cz5gtupa6y62h3q6a4gjv37pgefnpt5"
const faucetDefaultTokenAmount = "5000000"

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fund-faucet",
		Short: "Fund the Dymension faucet with rollapp tokens",
		Run: func(cmd *cobra.Command, args []string) {
			utils.PrettifyErrorIfExists(fundFaucet(cmd, args))
		},
	}
	cmd.Flags().StringP(flagNames.tokenAmount, "", faucetDefaultTokenAmount, "The amount of tokens to fund the faucet with")
	cmd.Flags().BoolP(flagNames.NoOutput, "", false, "Run the command without output")
	return cmd
}

func fundFaucet(cmd *cobra.Command, args []string) error {
	home := cmd.Flag(utils.FlagNames.Home).Value.String()
	rlpCfg, err := config.LoadConfigFromTOML(home)
	if err != nil {
		return err
	}
	noOutput, err := cmd.Flags().GetBool(flagNames.NoOutput)
	outputHandler := utils.NewOutputHandler(noOutput)
	if err != nil {
		return err
	}
	defer outputHandler.StopSpinner()
	utils.RunOnInterrupt(outputHandler.StopSpinner)
	outputHandler.StartSpinner("Getting relayer channel...")
	rly := relayer.NewRelayer(rlpCfg.Home, rlpCfg.RollappID, rlpCfg.HubData.ID)
	srcChannel, _, err := rly.LoadChannels()
	if err != nil || srcChannel == "" {
		return errors.New("failed to load relayer channel. Please make sure that the rollapp is " +
			"running on your local machine and a relayer channel has been established")
	}
	faucetTokenAmountStr := cmd.Flag(flagNames.tokenAmount).Value.String()
	faucetTokensAmount, success := new(big.Int).SetString(faucetTokenAmountStr, 10)
	if !success {
		return fmt.Errorf("invalid faucet %s token amount", faucetTokenAmountStr)
	}
	actualFaucetAmount := faucetTokensAmount.Mul(faucetTokensAmount, new(big.Int).Exp(big.NewInt(10),
		new(big.Int).SetUint64(uint64(rlpCfg.Decimals)), nil))
	fundFaucetCmd := exec.Command(rlpCfg.RollappBinary, "tx", "ibc-transfer", "transfer", "transfer",
		srcChannel, faucetAddr, actualFaucetAmount.String()+rlpCfg.Denom, "--from",
		consts.KeysIds.RollappSequencer, "--keyring-backend", "test", "--home", filepath.Join(rlpCfg.Home,
			consts.ConfigDirName.Rollapp), "--broadcast-mode", "block", "-y", "--output", "json")
	outputHandler.StartSpinner("Funding faucet...")
	stdout, err := utils.ExecBashCommandWithStdout(fundFaucetCmd)
	if err != nil {
		return err
	}
	err = tx_utils.CheckTxStdOut(stdout)
	if err != nil {
		return err
	}
	outputHandler.StopSpinner()
	outputHandler.DisplayMessage("💈 The Dymension faucet has been funded successfully!")
	return nil
}
