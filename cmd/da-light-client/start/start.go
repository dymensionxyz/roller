package start

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/roller"
)

const (
	rpcEndpointFlag     = "rpc-endpoint"
	metricsEndpointFlag = "metrics-endpoint"
)

func Cmd() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "start",
		Short: "Runs the DA light client.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String()
			pterm.Info.Println("loading roller config file")
			rollerData, err := roller.LoadConfig(home)
			errorhandling.PrettifyErrorIfExists(err)

			// TODO: refactor the version comparison for migrations
			// errorhandling.RequireMigrateIfNeeded(rollerData)

			metricsEndpoint := cmd.Flag(metricsEndpointFlag).Value.String()
			if metricsEndpoint != "" && rollerData.DA.Backend != consts.Celestia {
				errorhandling.PrettifyErrorIfExists(
					errors.New("metrics endpoint can only be set for celestia"),
				)
			}
			damanager := datalayer.NewDAManager(
				rollerData.DA.Backend,
				rollerData.Home,
				rollerData.KeyringBackend,
			)

			if rollerData.NodeType == "sequencer" {
				pterm.Info.Println("checking for da address balance")
				insufficientBalances, err := damanager.CheckDABalance()
				errorhandling.PrettifyErrorIfExists(err)
				err = keys.PrintInsufficientBalancesIfAny(insufficientBalances)
				if err != nil {
					pterm.Error.Println("failed to retrieve insufficient balances: ", err)
					return
				}
			}

			damanager.SetRPCEndpoint(rollerData.DA.CurrentStateNode)
			if metricsEndpoint != "" {
				damanager.SetMetricsEndpoint(metricsEndpoint)
			}

			startDALCCmd := damanager.GetStartDACmd()
			if startDALCCmd == nil {
				errorhandling.PrettifyErrorIfExists(
					errors.New(
						"DA doesn't need to run separately. It runs automatically with the app",
					),
				)
			}

			fmt.Println(startDALCCmd.String())
			done := make(chan error, 1)
			if rollerData.KeyringBackend == consts.SupportedKeyringBackends.OS {
				pswFileName, err := filesystem.GetOsKeyringPswFileName(consts.Executables.Celestia)
				if err != nil {
					pterm.Error.Println("failed to get os keyring password file name: ", err)
					return
				}

				fp := filepath.Join(home, string(pswFileName))
				psw, err := filesystem.ReadFromFile(fp)
				if err != nil {
					pterm.Error.Println("failed to read os keyring password file: ", err)
					return
				}

				pr := map[string]string{
					"Enter keyring passphrase":    psw,
					"Re-enter keyring passphrase": psw,
				}

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				// nolint: errcheck
				go func() {
					err := bash.ExecCmdFollow(
						done,
						ctx,
						startDALCCmd,
						pr, // No need for printOutput since we configured output above
					)

					done <- err
				}()

				select {
				case err := <-done:
					if err != nil {
						pterm.Error.Println("rollapp's process returned an error: ", err)
						return
					}
				case <-ctx.Done():
					pterm.Error.Println("context cancelled, terminating command")
					return
				}
			} else {
				ctx, cancel := context.WithCancel(cmd.Context())
				defer cancel()

				go func() {
					err := bash.ExecCmdFollow(
						done,
						ctx,
						startDALCCmd,
						nil,
					)

					done <- err
				}()

				select {
				case err := <-done:
					if err != nil {
						pterm.Error.Println("rollapp's process returned an error: ", err)
						return
					}
				case <-ctx.Done():
					pterm.Error.Println("context cancelled, terminating command")
					return
				}
			}
			select {}
		},
	}

	addFlags(runCmd)
	return runCmd
}

func addFlags(cmd *cobra.Command) {
	cmd.Flags().
		StringP(rpcEndpointFlag, "", "mocha-4-consensus.mesa.newmetric.xyz", "The DA rpc endpoint to connect to.")
	cmd.Flags().
		StringP(metricsEndpointFlag, "", "", "The OTEL collector metrics endpoint to connect to.")
}
