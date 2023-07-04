package register

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/dymensionxyz/roller/cmd/consts"

	"encoding/json"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"github.com/spf13/cobra"
)

// TODO: Test registration on 35-C and update the price
var registerUdymPrice = big.NewInt(1)

func Cmd() *cobra.Command {
	registerCmd := &cobra.Command{
		Use:   "register",
		Short: "Registers the rollapp and the sequencer to the Dymension hub.",
		Run: func(cmd *cobra.Command, args []string) {
			spin := utils.GetLoadingSpinner()
			spin.Suffix = consts.SpinnerMsgs.BalancesVerification
			spin.Start()
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rollappConfig, err := config.LoadConfigFromTOML(home)
			utils.PrettifyErrorIfExists(err)
			notFundedAddrs, err := utils.GetSequencerInsufficientAddrs(rollappConfig, registerUdymPrice)
			utils.PrettifyErrorIfExists(err)
			utils.PrintInsufficientBalancesIfAny(notFundedAddrs)
			spin.Suffix = consts.SpinnerMsgs.UniqueIdVerification
			spin.Restart()
			utils.PrettifyErrorIfExists(initconfig.VerifyUniqueRollappID(rollappConfig.RollappID, rollappConfig))
			spin.Suffix = " Registering RollApp to hub...\n"
			spin.Restart()
			utils.PrettifyErrorIfExists(registerRollapp(rollappConfig))
			registerSequencerCmd, err := getRegisterSequencerCmd(rollappConfig)
			utils.PrettifyErrorIfExists(err)
			spin.Suffix = " Registering RollApp sequencer...\n"
			spin.Restart()
			_, err = utils.ExecBashCommand(registerSequencerCmd)
			utils.PrettifyErrorIfExists(err)
			spin.Stop()
			printRegisterOutput(rollappConfig)
		},
	}

	return registerCmd
}

func registerRollapp(rollappConfig config.RollappConfig) error {
	cmd := getRegisterRollappCmd(rollappConfig)
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
	if err := handleStdOut(stdout, rollappConfig); err != nil {
		return err
	}
	return nil
}

type Response struct {
	RawLog string `json:"raw_log"`
}

func handleStdOut(stdout bytes.Buffer, rollappConfig config.RollappConfig) error {
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

func printRegisterOutput(rollappConfig config.RollappConfig) {
	fmt.Printf("💈 Rollapp '%s' has been successfully registered on the hub.\n", rollappConfig.RollappID)
}
