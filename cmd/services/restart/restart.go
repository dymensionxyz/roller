package restart

import (
	"slices"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/migrations"
	"github.com/dymensionxyz/roller/utils/roller"
	sequencerutils "github.com/dymensionxyz/roller/utils/sequencer"
	servicemanager "github.com/dymensionxyz/roller/utils/service_manager"
	"github.com/dymensionxyz/roller/utils/upgrades"
)

func Cmd(services []string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Restarts the systemd services relevant to RollApp",
		Run: func(cmd *cobra.Command, args []string) {
			home, err := filesystem.ExpandHomePath(
				cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String(),
			)
			if err != nil {
				pterm.Error.Println("failed to expand home directory")
				return
			}

			if slices.Contains(services, "rollapp") {
				rollappConfig, err := roller.LoadConfig(home)
				errorhandling.PrettifyErrorIfExists(err)

				if rollappConfig.NodeType == "sequencer" {
					err = sequencerutils.CheckBalance(rollappConfig)
					if err != nil {
						pterm.Error.Println("failed to check sequencer balance: ", err)
						return
					}

					warnErr := sequencerutils.WarnIfSequencerBelowLivenessSlashMin(rollappConfig)
					if warnErr != nil {
						pterm.Warning.Println("failed to evaluate liveness slash minimum:", warnErr)
					}
				}

				if rollappConfig.HubData.ID != consts.MockHubID {
					raUpgrade, err := upgrades.NewRollappUpgrade(
						string(rollappConfig.RollappVMType),
					)
					if err != nil {
						pterm.Error.Println("failed to check rollapp version equality: ", err)
					}

					err = migrations.RequireRollappMigrateIfNeeded(
						raUpgrade.CurrentVersionCommit[:6],
						rollappConfig.RollappBinaryVersion[:6],
						string(rollappConfig.RollappVMType),
					)
					if err != nil {
						pterm.Error.Println(err)
						return
					}
				}

				if slices.Contains(services, "rollapp") &&
					rollappConfig.DA.Backend == consts.Celestia {
					services = consts.RollappWithCelesSystemdServices
				}
			}

			err = servicemanager.RestartSystemServices(services, home)
			if err != nil {
				pterm.Error.Println("failed to restart systemd services:", err)
				return
			}
		},
	}
	return cmd
}
