package initrollapp

import (
	"fmt"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/config"
	"github.com/dymensionxyz/roller/utils/config/scripts"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/rollapp"
	"github.com/dymensionxyz/roller/utils/roller"
	"github.com/dymensionxyz/roller/utils/sequencer"
	servicemanager "github.com/dymensionxyz/roller/utils/service_manager"
	"github.com/dymensionxyz/roller/version"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [rollapp-id]",
		Short: "Initialize a RollApp configuration.",
		Long:  ``,
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := initconfig.AddFlags(cmd)
			if err != nil {
				pterm.Error.Println("failed to initialize rollapp: ", err)
				return
			}

			home, err := filesystem.ExpandHomePath(
				cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String(),
			)
			if err != nil {
				pterm.Error.Println("failed to initialize rollapp: ", err)
				return
			}

			shouldUseMockBackend, _ := cmd.Flags().GetBool("mock")
			shouldSkipBinaryInstallation, _ := cmd.Flags().GetBool("skip-binary-installation")

			// preflight checks
			var hd consts.HubData
			var env string
			var raID string

			isRootExist, err := filesystem.DirNotEmpty(home)
			if err != nil {
				pterm.Error.Printf(
					"failed to check if roller home directory (%s) is empty: %v\n",
					home,
					err,
				)
				return
			}

			if isRootExist {
				shouldContinue, err := sequencer.CheckExistingSequencer(home)
				if err != nil {
					pterm.Error.Printf(
						"failed to check if sequencer is already registered: %v\n",
						err,
					)
					return
				}

				if shouldContinue.IsSequencerAlreadyRegistered ||
					shouldContinue.IsSequencerProposer {
					pterm.Warning.Println("conditions to continue not met")
					yamlBytes, err := yaml.Marshal(shouldContinue)
					if err != nil {
						pterm.Error.Printf("failed to marshal sequencer address status: %v\n", err)
						return
					}

					fmt.Println(string(yamlBytes))

					pterm.Warning.Println("the existing hub_sequencer key is already registered")
					pterm.Warning.Println("start your rollapp instead")
					pterm.Warning.Println(
						"if you are resetting the node, remove the roller directory and run the command again",
					)
					return
				}
			}

			pterm.Info.Println("stopping system services for all component, if any...")
			err = servicemanager.StopSystemServices(consts.AllServices)
			if err != nil {
				pterm.Error.Println("failed to stop system services: ", err)
				return
			}

			err = filesystem.CreateRollerRootWithOptionalOverride(home)
			if err != nil {
				pterm.Error.Printf(
					"failed to create roller home directory (%s): %v\n",
					home,
					err,
				)
				return
			}

			isFirstInitialization, err := roller.CreateConfigFileIfNotPresent(home)
			if err != nil {
				pterm.Error.Println("failed to initialize rollapp: ", err)
				return
			}

			// get environment data
			if shouldUseMockBackend {
				env = "mock"
			} else {
				envs := []string{"mock", "playground", "blumbus", "custom", "mainnet"}
				env, _ = pterm.DefaultInteractiveSelect.
					WithDefaultText("select the environment you want to initialize for").
					WithOptions(envs).
					Show()
			}

			switch env {
			case "custom":
				chd, err := config.CreateCustomHubData()
				if err != nil {
					pterm.Error.Println("failed to create custom hub data: ", err)
					return
				}

				hd = consts.HubData{
					Environment:   env,
					ApiUrl:        chd.ApiUrl,
					ID:            chd.ID,
					RpcUrl:        chd.RpcUrl,
					ArchiveRpcUrl: chd.RpcUrl,
					GasPrice:      chd.GasPrice,
				}
			case "mock":
				// FIXME: handle mock (why it need custom handling? it's in consts)
				return
			default:
				hd = consts.Hubs[env]
			}

			kb := keys.KeyringBackendFromEnv(env)

			// if roller has not been initialized or it was reset
			// set the versions to the current version
			if isFirstInitialization {
				rollerConfigFilePath := roller.GetConfigPath(home)

				fieldsToUpdate := map[string]any{
					"roller_version":  version.BuildVersion,
					"keyring_backend": string(kb),
				}
				err = tomlconfig.UpdateFieldsInFile(rollerConfigFilePath, fieldsToUpdate)
				if err != nil {
					pterm.Error.Println("failed to update roller config file: ", err)
					return
				}
			}

			// get rollapp data
			if len(args) != 0 {
				raID = args[0]
			} else {
				raID, _ = pterm.DefaultInteractiveTextInput.WithDefaultText(
					"provide a rollapp ID that you want to run the node for",
				).Show()
			}
			raID = strings.TrimSpace(raID)
			_, err = rollapp.ValidateChainID(raID)
			if err != nil {
				pterm.Error.Println("failed to validate chain id: ", err)
				return
			}

			// default flow
			isRollappRegistered, _ := rollapp.IsRegistered(raID, hd)
			if !isRollappRegistered {
				pterm.Error.Printf("%s was not found as a registered rollapp: %v\n", raID, err)
				return
			}

			raResponse, err := rollapp.Show(raID, hd)
			if err != nil {
				pterm.Error.Println("failed to retrieve rollapp information: ", err)
				return
			}
			ra := raResponse.Rollapp

			if ra.GenesisInfo.Bech32Prefix == "" {
				pterm.Error.Println(
					"RollApp does not contain Bech32Prefix, which is mandatory to continue",
				)
				return
			}

			if !shouldSkipBinaryInstallation {
				start := time.Now()
				err := installDependencies(home, env, ra)
				if err != nil {
					pterm.Error.Println("failed to install dependencies: ", err)
					return
				}
				elapsed := time.Since(start)
				pterm.Info.Println("all dependencies installed in: ", elapsed)
			}

			// validate bech32 prefix in the installed binary
			bp, err := rollapp.ExtractBech32PrefixFromBinary(
				strings.ToLower(ra.VmType),
			)
			if err != nil {
				pterm.Error.Println("failed to extract bech32 prefix from binary", err)
				return
			}

			if ra.GenesisInfo.Bech32Prefix != bp {
				pterm.Error.Printf(
					"rollapp bech32 prefix does not match, want: %s, have: %s",
					ra.GenesisInfo.Bech32Prefix,
					bp,
				)
				return
			}

			// initialize rollapp node
			err = runInit(home, env, hd, *raResponse, kb)
			if err != nil {
				pterm.Error.Printf("failed to initialize the RollApp: %v\n", err)
				return
			}

			if kb == consts.SupportedKeyringBackends.OS {
				pterm.Info.Println("creating startup scripts for OS keyring backend")
				err := scripts.CreateRollappStartup(home)
				if err != nil {
					pterm.Error.Println("failed to generate startup scripts:", err)
					return
				}
			}

			defer func() {
				pterm.Info.Println("next steps:")
				pterm.Info.Printf(
					"run %s prepare node configuration for %s RollApp\n",
					pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
						Sprintf("roller rollapp setup"),
					raID,
				)
			}()
		},
	}

	cmd.Flags().Bool("mock", false, "initialize the rollapp with mock backend")
	cmd.Flags().Bool("skip-binary-installation", false, "skips the binary installation")

	return cmd
}
