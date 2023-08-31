package fund_faucet

import (
	"errors"
	"fmt"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"github.com/dymensionxyz/roller/relayer"
	"github.com/spf13/cobra"
	"os/exec"
	"path/filepath"
)

func Cmd() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "fund-faucet",
		Short: "Fund the Dymension faucet with rollapp tokens",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rlpCfg, err := config.LoadConfigFromTOML(home)
			utils.PrettifyErrorIfExists(err)
			rly := relayer.NewRelayer(rlpCfg.Home, rlpCfg.RollappID, rlpCfg.HubData.ID)
			srcChannel, _, err := rly.LoadChannels()
			if err != nil || srcChannel == "" {
				utils.PrettifyErrorIfExists(errors.New("failed to load relayer channel. Please make sure that the rollapp is " +
					"running on your local machine and a channel has been established"))
			}
			fundFuacetCmd := exec.Command(rlpCfg.RollappBinary, "tx", "ibc-transfer", "transfer", "transfer",
				srcChannel, "dym1g8sf7w4cz5gtupa6y62h3q6a4gjv37pgefnpt5", "5000000000000000000000000"+rlpCfg.Denom, "--from",
				consts.KeysIds.RollappSequencer, "--keyring-backend", "test", "--home", filepath.Join(rlpCfg.Home,
					consts.ConfigDirName.Rollapp), "--broadcast-mode", "block")
			stdout, err := utils.ExecBashCommandWithStdout(fundFuacetCmd)
			utils.PrettifyErrorIfExists(err)
			fmt.Println("💈 The Dymension faucet has been funded successfully!")
		},
	}
	return versionCmd
}
