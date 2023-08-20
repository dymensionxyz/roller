package register

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os/exec"
	"strings"

	"github.com/dymensionxyz/roller/cmd/consts"

	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"github.com/spf13/cobra"
)

// TODO: Test registration on 35-C and update the price
var registerUdymPrice = big.NewInt(1)

var flagNames = struct {
	NoOutput             string
	forceSeqRegistration string
}{
	NoOutput:             "no-output",
	forceSeqRegistration: "force",
}

func Cmd() *cobra.Command {
	registerCmd := &cobra.Command{
		Use:   "register",
		Short: "Registers the rollapp and the sequencer to the Dymension hub.",
		Run: func(cmd *cobra.Command, args []string) {
			utils.PrettifyErrorIfExists(register(cmd, args))
		},
	}

	registerCmd.Flags().BoolP(flagNames.forceSeqRegistration, "f", false, "force sequencer registration even if the rollapp is already registered")
	registerCmd.Flags().BoolP(flagNames.NoOutput, "", false, "Register the rollapp without output.")
	return registerCmd
}

func register(cmd *cobra.Command, args []string) error {
	noOutput, err := cmd.Flags().GetBool(flagNames.NoOutput)
	outputHandler := utils.NewOutputHandler(noOutput)
	if err != nil {
		return err
	}
	defer outputHandler.StopSpinner()
	utils.RunOnInterrupt(outputHandler.StopSpinner)
	outputHandler.StartSpinner(consts.SpinnerMsgs.BalancesVerification)
	home := cmd.Flag(utils.FlagNames.Home).Value.String()
	rollappConfig, err := config.LoadConfigFromTOML(home)
	if err != nil {
		return err
	}
	notFundedAddrs, err := utils.GetSequencerInsufficientAddrs(rollappConfig, registerUdymPrice)
	if err != nil {
		return err
	}
	if len(notFundedAddrs) > 0 {
		outputHandler.StopSpinner()
		utils.PrintInsufficientBalancesIfAny(notFundedAddrs, rollappConfig)
	}
	outputHandler.StartSpinner(" Registering RollApp to hub...\n")
	registerRollappCmd := getRegisterRollappCmd(rollappConfig)
	if err := runcCommandWithErrorChecking(registerRollappCmd); err != nil {
		if cmd.Flag(flagNames.forceSeqRegistration).Changed {
			fmt.Println("RollApp is already registered. Attempting to register sequencer anyway...")
		} else {
			return err
		}
	}
	registerSequencerCmd, err := getRegisterSequencerCmd(rollappConfig)
	if err != nil {
		return err
	}
	outputHandler.StartSpinner(" Registering RollApp sequencer...\n")
	err = runcCommandWithErrorChecking(registerSequencerCmd)
	if err != nil {
		return err
	}
	outputHandler.StopSpinner()
	outputHandler.DisplayMessage(fmt.Sprintf("💈 Rollapp '%s' has been successfully registered on the hub.", rollappConfig.RollappID))
	return nil
}

// TODO: probably should be moved to utils
func runcCommandWithErrorChecking(cmd *exec.Cmd) error {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmdExecErr := cmd.Run()
	if len(stderr.String()) > 0 {
		return errors.New(stderr.String())
	}
	if cmdExecErr != nil {
		return cmdExecErr
	}
	if err := handleStdOut(stdout); err != nil {
		return err
	}
	return nil
}

type Response struct {
	RawLog string `json:"raw_log"`
}

func handleStdOut(stdout bytes.Buffer) error {
	var response Response

	err := json.NewDecoder(&stdout).Decode(&response)
	if err != nil {
		return err
	}

	if strings.Contains(response.RawLog, "fail") || strings.Contains(response.RawLog, "error") {
		return errors.New(response.RawLog)
	}

	return nil
}
