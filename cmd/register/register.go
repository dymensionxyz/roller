package register

import (
	"bytes"
	"errors"
	"os/exec"
	"path/filepath"

	"fmt"

	"strings"

	"encoding/json"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/spf13/cobra"
)

func RegisterCmd() *cobra.Command {
	registerCmd := &cobra.Command{
		Use:   "register",
		Short: "Registers the rollapp and the sequencer to the Dymension hub.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(initconfig.FlagNames.Home).Value.String()
			rollappConfig, err := initconfig.LoadConfigFromTOML(home)
			initconfig.OutputCleanError(err)
			initconfig.OutputCleanError(initconfig.VerifyUniqueRollappID(rollappConfig.RollappID, rollappConfig))
			initconfig.OutputCleanError(registerRollapp(rollappConfig))
			printRegisterOutput(rollappConfig)
		},
	}
	addFlags(registerCmd)
	return registerCmd
}

func addFlags(cmd *cobra.Command) {
	cmd.Flags().StringP(initconfig.FlagNames.Home, "", initconfig.GetRollerRootDir(), "The directory of the roller config files")
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
	if err := handleStdOut(stdout, rollappConfig); err != nil {
		return err
	}
	if cmdExecErr != nil {
		return cmdExecErr
	}
	return nil
}

func handleStdErr(stderr bytes.Buffer, rollappConfig initconfig.InitConfig) error {
	stderrStr := stderr.String()
	if len(stderrStr) > 0 {
		if strings.Contains(stderrStr, "key not found") {
			sequencerAddress, err := initconfig.GetAddress(
				initconfig.KeyConfig{
					ID:       initconfig.KeyNames.HubSequencer,
					Prefix:   initconfig.AddressPrefixes.Hub,
					Dir:      filepath.Join(rollappConfig.Home, initconfig.ConfigDirName.Rollapp),
					CoinType: initconfig.CoinTypes.Cosmos,
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
		"--from", initconfig.KeyNames.HubSequencer,
		"--keyring-backend", "test",
		"--keyring-dir", filepath.Join(rollappConfig.Home, initconfig.ConfigDirName.Rollapp),
		rollappConfig.RollappID, "stamp1", "genesis-path/1", "3", "3", `{"Addresses":[]}`, "--output", "json",
		"--node", rollappConfig.HubData.RPC_URL, "--yes", "--broadcast-mode", "block", "--chain-id", rollappConfig.HubData.ID,
	)
}

func printRegisterOutput(rollappConfig initconfig.InitConfig) {
	fmt.Printf("💈 Rollapp '%s' has been successfully registered on the hub.\n", rollappConfig.RollappID)
}
