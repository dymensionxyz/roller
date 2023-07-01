package export

import (
	"fmt"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/spf13/cobra"
	"os/exec"
	"path/filepath"
	"strings"
)

var supportedKeys = []string{
	consts.KeysIds.HubSequencer,
	consts.KeysIds.RollappSequencer,
}

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "export <key-id>",
		Short: fmt.Sprintf("Exports the private key of the given key id. The supported keys are %s",
			strings.Join(supportedKeys, ", ")),
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			config, err := utils.LoadConfigFromTOML(home)
			utils.PrettifyErrorIfExists(err)

			keyID := args[0]
			if keyID == consts.KeysIds.HubSequencer {
				exportKeyCmd := getExportKeyCmdBinary(keyID, filepath.Join(home, consts.ConfigDirName.HubKeys),
					consts.Executables.Dymension)
				out, err := utils.ExecBashCommand(exportKeyCmd)
				utils.PrettifyErrorIfExists(err)
				fmt.Println(out.String())
			} else if keyID == consts.KeysIds.RollappSequencer {
				exportKeyCmd := getExportKeyCmdBinary(keyID, filepath.Join(home, consts.ConfigDirName.Rollapp),
					config.RollappBinary)
				out, err := utils.ExecBashCommand(exportKeyCmd)
				utils.PrettifyErrorIfExists(err)
				fmt.Println(out.String())
			} else {
				utils.PrettifyErrorIfExists(fmt.Errorf("invalid key id: %s. The supported keys are %s", keyID,
					strings.Join(supportedKeys, ", ")))
			}
		},
		Args: cobra.ExactArgs(1),
	}
	utils.AddGlobalFlags(cmd)
	return cmd
}

func getExportKeyCmdBinary(keyID, keyringDir, binary string) *exec.Cmd {
	flags := getExportKeyFlags(keyringDir)
	cmdStr := fmt.Sprintf("yes | %s keys export %s %s", binary, keyID, flags)
	return exec.Command("bash", "-c", cmdStr)
}

func getExportKeyFlags(keyringDir string) string {
	return fmt.Sprintf("--keyring-backend test --keyring-dir %s --unarmored-hex --unsafe", keyringDir)
}
