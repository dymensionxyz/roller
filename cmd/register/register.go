package register

import (
	"bytes"
	"os/exec"
	"path/filepath"

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
			initconfig.OutputCleanError(registerRollapp(rollappConfig))
		},
	}
	addFlags(registerCmd)
	return registerCmd
}

func addFlags(cmd *cobra.Command) {
	cmd.Flags().StringP(initconfig.FlagNames.Home, "", initconfig.GetRollerRootDir(), "The directory of the roller config files")
}

func registerRollapp(rollappConfig initconfig.InitConfig) error {
	cmd := exec.Command(
		consts.Executables.Dymension, "tx", "rollapp", "create-rollapp",
		"--from", initconfig.KeyNames.HubSequencer,
		"--keyring-backend", "test",
		"--keyring-dir", filepath.Join(rollappConfig.Home, initconfig.ConfigDirName.Rollapp),
		rollappConfig.RollappID, "stamp1", "genesis-path/1", "3", "3", `{"Addresses":[]}`, "--output", "json",
		"--node", initconfig.HubData.RPC_URL, "--yes", "--broadcast-mode", "block",
	)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	error := cmd.Run()
	return error
}
