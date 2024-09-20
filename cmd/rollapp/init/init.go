package initrollapp

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/dependencies"
	"github.com/dymensionxyz/roller/utils/dependencies/types"
	"github.com/dymensionxyz/roller/utils/rollapp"
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
				pterm.Error.Println("failed to add flags")
				return
			}
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			isMockFlagSet := cmd.Flags().Changed("mock")
			shouldUseMockBackend, _ := cmd.Flags().GetBool("mock")
			// rollerData, err := tomlconfig.LoadRollerConfig(home)
			fmt.Println(home)
			// TODO: move to consts
			// TODO(v2):  move to roller config
			if !shouldUseMockBackend {
				dymdBinaryOptions := types.Dependency{
					Name:       "dymension",
					Repository: "https://github.com/artemijspavlovs/dymension",
					Release:    "v3.1.0-pg07",
					Binaries: []types.BinaryPathPair{
						{
							Binary:            "dymd",
							BinaryDestination: consts.Executables.Dymension,
							BuildCommand: exec.Command(
								"make",
								"build",
							),
						},
					},
				}
				pterm.Info.Println("installing dependencies")
				err = dependencies.InstallBinaryFromRelease(dymdBinaryOptions)
				if err != nil {
					pterm.Error.Println("failed to install dymd: ", err)
					return
				}

			}

			var hd consts.HubData
			var env string
			var raID string

			if shouldUseMockBackend {
				env = "mock"
			}

			if !isMockFlagSet && !shouldUseMockBackend {
				envs := []string{"mock", "playground"}
				env, _ = pterm.DefaultInteractiveSelect.
					WithDefaultText("select the environment you want to initialize for").
					WithOptions(envs).
					Show()
			}
			hd = consts.Hubs[env]

			if len(args) != 0 {
				raID = args[0]
			} else {
				raID, _ = pterm.DefaultInteractiveTextInput.WithDefaultText(
					"provide a rollapp ID that you want to run the node for",
				).Show()
			}

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
				err = dependencies.InstallBinaries(home, true, raRespMock)
				if err != nil {
					pterm.Error.Println("failed to install binaries: ", err)
					return
				}
				err := runInit(
					cmd,
					env,
					raRespMock,
				)
				if err != nil {
					fmt.Println("failed to run init: ", err)
					return
				}
				return
			}

			// ex binaries install

			raID = strings.TrimSpace(raID)

			getRaCmd := rollapp.GetRollappCmd(raID, hd)
			var raResponse rollapp.ShowRollappResponse

			out, err := bash.ExecCommandWithStdout(getRaCmd)
			if err != nil {
				pterm.Error.Println("failed to get rollapp: ", err)
				return
			}
			err = json.Unmarshal(out.Bytes(), &raResponse)
			if err != nil {
				pterm.Error.Println("failed to unmarshal", err)
				return
			}

			if raResponse.Rollapp.GenesisInfo.Bech32Prefix == "" {
				pterm.Error.Println("no bech")
				return
			}
			start := time.Now()
			err = dependencies.InstallBinaries(home, false, raResponse)
			if err != nil {
				pterm.Error.Println("failed to install binaries: ", err)
				return
			}
			elapsed := time.Since(start)

			pterm.Info.Println("all dependencies installed in: ", elapsed)
			// END: ex binaries install

			isRollappRegistered, _ := rollapp.IsRollappRegistered(raID, hd)
			if !isRollappRegistered {
				pterm.Error.Printf("%s was not found as a registered rollapp: %v", raID, err)
				return
			}

			bp, err := rollapp.ExtractBech32Prefix(
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

			err = runInit(cmd, env, raResponse)
			if err != nil {
				pterm.Error.Printf("failed to initialize the RollApp: %v\n", err)
				return
			}

			pterm.Info.Println("next steps:")
			pterm.Info.Printf(
				"run %s prepare node configuration for %s RollApp\n",
				pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
					Sprintf("roller rollapp setup"),
				raID,
			)
		},
	}

	cmd.Flags().Bool("mock", false, "initialize the rollapp with mock backend")

	return cmd
}
