package claim

import (
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/spf13/cobra"
	"os/exec"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim-rewards <private-key> <destination address>",
		Short: "Send the DYM rewards associated with the given private key to the destination address",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Replace with mainnet hub RPC
			//hubRPC := consts.Hubs[consts.FroopylandHubName].RPC_URL
			//tempDir, err := os.MkdirTemp(os.TempDir(), "hub_sequencer")
			//if err != nil {
			//	return err
			//}
			importKeyCmd := exec.Command(consts.Executables.Simd, "keys", "import-hex",
				consts.KeysIds.HubSequencer, args[0])
			_, err := utils.ExecBashCommandWithStdout(importKeyCmd)
			if err != nil {
				return err
			}
			return nil
		},
	}
	return cmd
}
