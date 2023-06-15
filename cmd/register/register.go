package register

import (
	"bytes"
	"errors"
	"github.com/dymensionxyz/roller/cmd/consts"
	"path/filepath"

	"fmt"

	"strings"

	"encoding/json"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/spf13/cobra"
)

func RegisterCmd() *cobra.Command {
	registerCmd := &cobra.Command{
		Use:   "register",
		Short: "Registers the rollapp and the sequencer to the Dymension hub.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(initconfig.FlagNames.Home).Value.String()
			rollappConfig, err := utils.LoadConfigFromTOML(home)
			utils.PrettifyErrorIfExists(err)
			utils.PrettifyErrorIfExists(initconfig.VerifyUniqueRollappID(rollappConfig.RollappID, rollappConfig))
			err = registerRollapp(rollappConfig)
			utils.PrettifyErrorIfExists(err)
			registerSequencerCmd, err := getRegisterSequencerCmd(rollappConfig)
			utils.PrettifyErrorIfExists(err)
			err = registerSequencerCmd.Run()
			utils.PrettifyErrorIfExists(err)
			printRegisterOutput(rollappConfig)
		},
	}
	utils.AddGlobalFlags(registerCmd)
	return registerCmd
}

func registerRollapp(rollappConfig utils.RollappConfig) error {
	cmd := getRegisterRollappCmd(rollappConfig)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmdExecErr := cmd.Run()
	if err := handleStdErr(stderr, rollappConfig); err != nil {
		return err
	}
	if cmdExecErr != nil {
		return cmdExecErr
	}
	if err := handleStdOut(stdout, rollappConfig); err != nil {
		return err
	}
	return nil
}

func handleStdErr(stderr bytes.Buffer, rollappConfig utils.RollappConfig) error {
	stderrStr := stderr.String()
	if len(stderrStr) > 0 {
		if strings.Contains(stderrStr, "key not found") {
			sequencerAddress, err := utils.GetAddress(
				utils.KeyConfig{
					ID:       consts.KeyNames.HubSequencer,
					Prefix:   consts.AddressPrefixes.Hub,
					Dir:      filepath.Join(rollappConfig.Home, consts.ConfigDirName.Rollapp),
					CoinType: consts.CoinTypes.Cosmos,
				},
			)
			if err != nil {
				return err
			}
			return fmt.Errorf("Insufficient funds in the sequencer's address to register the RollApp. Please deposit DYM to the following address: %s and attempt the registration again", sequencerAddress)
		}
		return errors.New(stderrStr)
	}
	return nil
}

type Response struct {
	RawLog string `json:"raw_log"`
}

func handleStdOut(stdout bytes.Buffer, rollappConfig utils.RollappConfig) error {
	var response Response

	err := json.NewDecoder(&stdout).Decode(&response)
	if err != nil {
		return err
	}

	if strings.Contains(response.RawLog, "fail") {
		return errors.New(response.RawLog)
	}

	return nil
}

func printRegisterOutput(rollappConfig utils.RollappConfig) {
	fmt.Printf("💈 Rollapp '%s' has been successfully registered on the hub.\n", rollappConfig.RollappID)
}
