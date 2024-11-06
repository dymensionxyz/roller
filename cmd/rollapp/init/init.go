package initrollapp

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/config"
	"github.com/dymensionxyz/roller/utils/config/scripts"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/dependencies"
	"github.com/dymensionxyz/roller/utils/dependencies/types"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/rollapp"
	"github.com/dymensionxyz/roller/utils/roller"
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

			err = servicemanager.StopSystemServices()
			if err != nil {
				pterm.Error.Println("failed to stop system services: ", err)
				return
			}

			err = filesystem.CreateDirWithOptionalOverwrite(home)
			if err != nil {
				pterm.Error.Println("failed to create roller home directory: ", err)
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
				envs := []string{"mock", "playground", "custom"}
				env, _ = pterm.DefaultInteractiveSelect.
					WithDefaultText("select the environment you want to initialize for").
					WithOptions(envs).
					Show()
			}

			// TODO: move to consts
			// TODO(v2):  move to roller config
			if !shouldUseMockBackend && env != "custom" {
				dymdBinaryOptions := types.Dependency{
					DependencyName:  "dymension",
					RepositoryOwner: "dymensionxyz",
					RepositoryName:  "dymension",
					RepositoryUrl:   "https://github.com/artemijspavlovs/dymension",
					Release:         "v3.1.0-pg10",
					Binaries: []types.BinaryPathPair{
						{
							Binary:            "dymd",
							BinaryDestination: consts.Executables.Dymension,
							BuildCommand:      exec.Command("make", "build"),
						},
					},
					PersistFiles: []types.PersistFile{},
				}
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
				hd = *chd

				if err != nil {
					pterm.Info.Println("failed to create custom hub data", err)
					return
				}

				err = dependencies.InstallCustomDymdVersion()
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
					_, _, err = dependencies.InstallBinaries(true, raRespMock)
					if err != nil {
						pterm.Error.Println("failed to install binaries: ", err)
						return
					}
				}

				err := runInit(
					cmd,
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
			builtDeps, _, err := dependencies.InstallBinaries(false, *raResponse)
			if err != nil {
				pterm.Error.Println("failed to install binaries: ", err)
				return
			}
			elapsed := time.Since(start)

			pterm.Info.Println("all dependencies installed in: ", elapsed)

			// if roller has not been initialized or it was reset
			// set the versions to the current version
			if isFirstInitialization {
				rollerConfigFilePath := roller.GetConfigPath(home)

				fieldsToUpdate := map[string]any{
					"roller_version":         version.BuildVersion,
					"rollapp_binary_version": builtDeps["rollapp"].Release,
					"keyring_backend":        kb,
				}
				err = tomlconfig.UpdateFieldsInFile(rollerConfigFilePath, fieldsToUpdate)
				if err != nil {
					pterm.Error.Println("failed to update roller config file: ", err)
					return
				}
			}

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

			err = runInit(cmd, env, hd, *raResponse, kb)
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
