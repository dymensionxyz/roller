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
	"github.com/dymensionxyz/roller/utils/dependencies"
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
			shouldUseMockBackend, _ := cmd.Flags().GetBool("mock")
			shouldSkipBinaryInstallation, _ := cmd.Flags().GetBool("skip-binary-installation")

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

			if shouldUseMockBackend {
				env = "mock"
			} else {
				envs := []string{"mock", "playground", "blumbus", "custom", "mainnet"}
				env, _ = pterm.DefaultInteractiveSelect.
					WithDefaultText("select the environment you want to initialize for").
					WithOptions(envs).
					Show()
			}

			if env == "mainnet" {
				pterm.Warning.Println(
					"By default roller uses a public endpoint which is not reliable. for production usage it's highly recommended to use a private endpoint. A freemium private endpoint can be obtained in the following link https://blastapi.io/chains/dymension",
				)
				pterm.Info.Printf(
					"run %s to update the Hub private endpoints anytime after initial setup\n",
					pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
						Sprintf("roller config set hub-rpc-endpoint <private-endpoint>"),
				)
			}

			// TODO: move to consts
			// TODO(v2):  move to roller config
			if !shouldUseMockBackend && env != "custom" {
				dymdBinaryOptions := dependencies.DefaultDymdDependency()
				pterm.Info.Println("installing dependencies")
				err = dependencies.InstallBinaryFromRelease(dymdBinaryOptions)
				if err != nil {
					pterm.Error.Println("failed to install dymd: ", err)
					return
				}
			}

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

			// env handling
			kb := keys.KeyringBackendFromEnv(env)
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

				err = dependencies.InstallCustomDymdVersion(chd.DymensionHash)
				if err != nil {
					pterm.Info.Println("failed to install dymd", err)
					return
				}
			case "mock":
				vmType := config.PromptVmType()
				raRespMock := rollapp.ShowRollappResponse{
					Rollapp: rollapp.Rollapp{
						RollappId: raID,
						VmType:    vmType,
					},
				}

				if !shouldSkipBinaryInstallation {
					_, _, err = dependencies.InstallBinaries(true, raRespMock, env)
					if err != nil {
						pterm.Error.Println("failed to install binaries: ", err)
						return
					}
				}

				err := runInit(
					home,
					env,
					consts.HubData{},
					raRespMock,
					kb,
				)
				if err != nil {
					fmt.Println("failed to run init: ", err)
					return
				}
				return
			default:
				hd = consts.Hubs[env]

				if shouldSkipBinaryInstallation {
					dymdDep := dependencies.DefaultDymdDependency()
					err = dependencies.InstallBinaryFromRelease(dymdDep)
					if err != nil {
						pterm.Error.Println("failed to install dymd: ", err)
						return
					}
				}
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

			if raResponse.Rollapp.GenesisInfo.Bech32Prefix == "" {
				pterm.Error.Println(
					"RollApp does not contain Bech32Prefix, which is mandatory to continue",
				)
				return
			}

			start := time.Now()
			builtDeps, _, err := dependencies.InstallBinaries(false, *raResponse, env)
			if err != nil {
				pterm.Error.Println("failed to install binaries: ", err)
				return
			}
			elapsed := time.Since(start)

			pterm.Info.Println("all dependencies installed in: ", elapsed)

			bp, err := rollapp.ExtractBech32PrefixFromBinary(
				strings.ToLower(raResponse.Rollapp.VmType),
			)
			if err != nil {
				pterm.Error.Println("failed to extract bech32 prefix from binary", err)
			}

			if raResponse.Rollapp.GenesisInfo.Bech32Prefix != bp {
				pterm.Error.Printf(
					"rollapp bech32 prefix does not match, want: %s, have: %s",
					raResponse.Rollapp.GenesisInfo.Bech32Prefix,
					bp,
				)
				return
			}

			// TODO: all above should be wrapped in "InitDependencies"

			err = runInit(home, env, hd, *raResponse, kb)
			if err != nil {
				pterm.Error.Printf("failed to initialize the RollApp: %v\n", err)
				return
			}

			// if roller has not been initialized or it was reset
			// set the versions to the current version
			if isFirstInitialization {
				rollerConfigFilePath := roller.GetConfigPath(home)

				fieldsToUpdate := map[string]any{
					"roller_version":         version.BuildVersion,
					"rollapp_binary_version": builtDeps["rollapp"].Release,
					"keyring_backend":        string(kb),
				}

				err = tomlconfig.UpdateFieldsInFile(rollerConfigFilePath, fieldsToUpdate)
				if err != nil {
					pterm.Error.Println("failed to update roller config file: ", err)
					return
				}
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
