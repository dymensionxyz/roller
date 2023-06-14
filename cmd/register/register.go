package register

import (
	"bytes"
	"errors"
	"os/exec"
	"path/filepath"

	"fmt"

	"encoding/json"
	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/spf13/cobra"
	"strings"
)

func RegisterCmd() *cobra.Command {
	registerCmd := &cobra.Command{
		Use:   "register",
		Short: "Registers the rollapp and the sequencer to the Dymension hub.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(initconfig.FlagNames.Home).Value.String()
			rollappConfig, err := initconfig.LoadConfigFromTOML(home)
			utils.PrettifyErrorIfExists(err)
			utils.PrettifyErrorIfExists(initconfig.VerifyUniqueRollappID(rollappConfig.RollappID, rollappConfig))
			utils.PrettifyErrorIfExists(registerRollapp(rollappConfig))
			printRegisterOutput(rollappConfig)
		},
	}
	addFlags(registerCmd)
	return registerCmd
}

func addFlags(cmd *cobra.Command) {
	cmd.Flags().StringP(initconfig.FlagNames.Home, "", utils.GetRollerRootDir(), "The directory of the roller config files")
}

func registerRollapp(rollappConfig initconfig.InitConfig) error {
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

func handleStdErr(stderr bytes.Buffer, rollappConfig initconfig.InitConfig) error {
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

func handleStdOut(stdout bytes.Buffer, rollappConfig initconfig.InitConfig) error {
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

func getRegisterRollappCmd(rollappConfig initconfig.InitConfig) *exec.Cmd {
	return exec.Command(
		consts.Executables.Dymension, "tx", "rollapp", "create-rollapp",
		"--from", consts.KeyNames.HubSequencer,
		"--keyring-backend", "test",
		"--keyring-dir", filepath.Join(rollappConfig.Home, consts.ConfigDirName.Rollapp),
		rollappConfig.RollappID, "stamp1", "genesis-path/1", "3", "3", `{"Addresses":[]}`, "--output", "json",
		"--node", rollappConfig.HubData.RPC_URL, "--yes", "--broadcast-mode", "block", "--chain-id", rollappConfig.HubData.ID,
	)
}

func printRegisterOutput(rollappConfig initconfig.InitConfig) {
	fmt.Printf("💈 Rollapp '%s' has been successfully registered on the hub.\n", rollappConfig.RollappID)
}
