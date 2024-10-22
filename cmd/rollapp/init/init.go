package initrollapp

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/dymensionxyz/roller/utils/config"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/dependencies"
	"github.com/dymensionxyz/roller/utils/dependencies/types"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/rollapp"
	"github.com/dymensionxyz/roller/utils/roller"
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

			isMockFlagSet := cmd.Flags().Changed("mock")
			shouldUseMockBackend, _ := cmd.Flags().GetBool("mock")

			// check whether roller was already initialized on the host
			err = filesystem.CreateDirWithOptionalOverwrite(home)
			if err != nil {
				pterm.Error.Println("failed to create roller home directory: ", err)
				return
			}

			isFirstInitialization, err := roller.CreateConfigFile(home)
			if err != nil {
				pterm.Error.Println("failed to initialize rollapp: ", err)
				return
			}

			var hd consts.HubData
			var env string
			var raID string

			if shouldUseMockBackend {
				env = "mock"
			}

			if !isMockFlagSet && !shouldUseMockBackend {
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
					Release:         "v3.1.0-pg07",
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

			if env != "custom" {
				hd = consts.Hubs[env]
			} else {
				hd = config.GenerateCustomHubData()

				err = dependencies.InstallCustomDymdVersion()
				if err != nil {
					pterm.Error.Println("failed to install custom dymd version: ", err)
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

			if env == "mock" {
				vmtypes := []string{"evm", "wasm"}
				vmtype, _ := pterm.DefaultInteractiveSelect.
					WithDefaultText("select the rollapp VM type you want to initialize for").
					WithOptions(vmtypes).
					Show()
				raRespMock := rollapp.ShowRollappResponse{
					Rollapp: rollapp.Rollapp{
						RollappId: raID,
						VmType:    vmtype,
					},
				}

				_, _, err = dependencies.InstallBinaries(true, raRespMock)
				if err != nil {
					pterm.Error.Println("failed to install binaries: ", err)
					return
				}
				err := runInit(
					cmd,
					env,
					consts.HubData{},
					raRespMock,
				)
				if err != nil {
					fmt.Println("failed to run init: ", err)
					return
				}
				return
			}

			isRollappRegistered, _ := rollapp.IsRollappRegistered(raID, hd)
			if !isRollappRegistered {
				pterm.Error.Printf("%s was not found as a registered rollapp: %v", raID, err)
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

				valuesToUpdate := map[string]string{
					"roller_version":         version.BuildVersion,
					"rollapp_binary_version": builtDeps["rollapp"].Release,
				}

				for k, v := range valuesToUpdate {
					err := tomlconfig.UpdateFieldInFile(
						rollerConfigFilePath,
						k,
						v,
					)
					if err != nil {
						pterm.Error.Println("failed to update roller config file: ", err)
						return
					}
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

			err = runInit(cmd, env, hd, *raResponse)
			if err != nil {
				pterm.Error.Printf("failed to initialize the RollApp: %v\n", err)
				return
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

	return cmd
}
