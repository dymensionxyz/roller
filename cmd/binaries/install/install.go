package install

//
// import (
// 	"encoding/json"
// 	"fmt"
// 	"os/exec"
// 	"strings"
// 	"time"
//
// 	"github.com/dymensionxyz/roller/cmd/consts"
// 	"github.com/dymensionxyz/roller/utils/bash"
// 	"github.com/dymensionxyz/roller/utils/dependencies"
// 	"github.com/dymensionxyz/roller/utils/rollapp"
// 	"github.com/pterm/pterm"
// 	"github.com/spf13/cobra"
// )
//
// func Cmd() *cobra.Command {
// 	cmd := &cobra.Command{
// 		Use:   "install <rollapp-id>",
// 		Short: "Install necessary binaries for operating a RollApp node",
// 		Args:  cobra.MaximumNArgs(1),
// 		Run: func(cmd *cobra.Command, args []string) {
// 			// TODO: instead of relying on dymd binary, query the rpc for rollapp
// 			envs := []string{"playground"}
// 			env, _ := pterm.DefaultInteractiveSelect.
// 				WithDefaultText("select the environment you want to initialize for").
// 				WithOptions(envs).
// 				Show()
// 			hd := consts.Hubs[env]
//
// 			var raID string
// 			if len(args) != 0 {
// 				raID = args[0]
// 			} else {
// 				raID, _ = pterm.DefaultInteractiveTextInput.WithDefaultText(
// 					"provide RollApp ID you plan to run the nodes for",
// 				).Show()
// 			}
//
// 			dymdBinaryOptions := dependencies.Dependency{
// 				Name:       "dymension",
// 				Repository: "https://github.com/artemijspavlovs/dymension",
// 				Release:    "3.1.0-pg04",
// 				Binaries: []dependencies.BinaryPathPair{
// 					{
// 						Binary:            "dymd",
// 						BinaryDestination: consts.Executables.Dymension,
// 						BuildCommand: exec.Command(
// 							"make",
// 							"build",
// 						),
// 					},
// 				},
// 			}
//
// 			pterm.Info.Println("installing dependencies")
// 			err := installBinaryFromRelease(dymdBinaryOptions)
// 			if err != nil {
// 				return
// 			}
//
// 			raID = strings.TrimSpace(raID)
//
// 			getRaCmd := rollapp.GetRollappCmd(raID, hd)
// 			var raResponse rollapp.ShowRollappResponse
// 			out, err := bash.ExecCommandWithStdout(getRaCmd)
// 			if err != nil {
// 				pterm.Error.Println("failed to get rollapp: ", err)
// 				return
// 			}
//
// 			err = json.Unmarshal(out.Bytes(), &raResponse)
// 			if err != nil {
// 				pterm.Error.Println("failed to unmarshal", err)
// 				return
// 			}
//
// 			start := time.Now()
// 			if raResponse.Rollapp.GenesisInfo.Bech32Prefix == "" {
// 				pterm.Error.Println("no bech")
// 				return
// 			}
// 			installBinaries(raResponse.Rollapp.GenesisInfo.Bech32Prefix)
// 			elapsed := time.Since(start)
// 			fmt.Println("time elapsed: ", elapsed)
// 		},
// 	}
//
// 	cmd.Flags().String("node", consts.PlaygroundHubData.RPC_URL, "hub rpc endpoint")
// 	cmd.Flags().String("chain-id", consts.PlaygroundHubData.ID, "hub chain id")
//
// 	return cmd
// }
