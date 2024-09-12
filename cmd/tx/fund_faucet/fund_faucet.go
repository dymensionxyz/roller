package fund_faucet

import (
	"errors"
	"fmt"
	"math/big"
	"os/exec"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/tx/tx_utils"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/relayer"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/spf13/cobra"
)

var flagNames = struct {
	tokenAmount string
	NoOutput    string
}{
	tokenAmount: "token-amount",
	NoOutput:    "no-output",
}

const (
	faucetAddr               = "dym1g8sf7w4cz5gtupa6y62h3q6a4gjv37pgefnpt5"
	faucetDefaultTokenAmount = "5000000"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fund-faucet",
		Short: "Fund the Dymension faucet with rollapp tokens",
		Run: func(cmd *cobra.Command, args []string) {
			errorhandling.PrettifyErrorIfExists(fundFaucet(cmd, args))
		},
	}
	cmd.Flags().
		StringP(flagNames.tokenAmount, "", faucetDefaultTokenAmount, "The amount of tokens to fund the faucet with")
	cmd.Flags().BoolP(flagNames.NoOutput, "", false, "Run the command without output")
	return cmd
}

func fundFaucet(cmd *cobra.Command, args []string) error {
	home := cmd.Flag(utils.FlagNames.Home).Value.String()
	rlpCfg, err := tomlconfig.LoadRollerConfig(home)
	if err != nil {
		return err
	}
	noOutput, err := cmd.Flags().GetBool(flagNames.NoOutput)
	outputHandler := utils.NewOutputHandler(noOutput)
	if err != nil {
		return err
	}
	defer outputHandler.StopSpinner()
	errorhandling.RunOnInterrupt(outputHandler.StopSpinner)
	outputHandler.StartSpinner(" Loading relayer channel...")
	rly := relayer.NewRelayer(rlpCfg.Home, rlpCfg.RollappID, rlpCfg.HubData.ID)
	_, _, err = rly.LoadActiveChannel()
	if err != nil || rly.SrcChannel == "" {
		return errors.New(
			"failed to load relayer channel. Please make sure that the rollapp is " +
				"running on your local machine and a relayer channel has been established",
		)
	}
	faucetTokenAmountStr := cmd.Flag(flagNames.tokenAmount).Value.String()
	faucetTokensAmount, success := new(big.Int).SetString(faucetTokenAmountStr, 10)
	if !success {
		return fmt.Errorf("invalid faucet %s token amount", faucetTokenAmountStr)
	}
	actualFaucetAmount := faucetTokensAmount.Mul(
		faucetTokensAmount,
		new(big.Int).Exp(
			big.NewInt(10),
			new(big.Int).SetUint64(uint64(rlpCfg.Decimals)), nil,
		),
	)
	fundFaucetCmd := exec.Command(
		rlpCfg.RollappBinary,
		"tx",
		"ibc-transfer",
		"transfer",
		"transfer",
		rly.SrcChannel,
		faucetAddr,
		actualFaucetAmount.String()+rlpCfg.Denom,
		"--from",
		consts.KeysIds.RollappSequencer,
		"--keyring-backend",
		"test",
		"--home",
		filepath.Join(
			rlpCfg.Home,
			consts.ConfigDirName.Rollapp,
		),
		"--broadcast-mode",
		"block",
		"-y",
		"--output",
		"json",
	)
	outputHandler.StartSpinner(" Funding faucet...")
	stdout, err := bash.ExecCommandWithStdout(fundFaucetCmd)
	if err != nil {
		return err
	}
	err = tx_utils.CheckTxJsonStdOut(stdout)
	if err != nil {
		return err
	}
	outputHandler.StopSpinner()
	outputHandler.DisplayMessage("💈 The Dymension faucet has been funded successfully!")
	return nil
}
