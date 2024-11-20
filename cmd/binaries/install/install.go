package install

import (
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/consts"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install <rollapp-id>",
		Short: "Install necessary binaries for operating a RollApp node",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			pterm.Info.Println("not implemented")
			// home := cmd.Flag(utils.FlagNames.Home).Value.String()
			//
			// rollerData, err := tomlconfig.LoadRollerConfig(home)
			// if err != nil {
			// 	pterm.Error.Println("failed to load roller config file", err)
			// 	return
			// }
			//
			// // TODO: instead of relying on dymd binary, query the rpc for rollapp
			// envs := []string{"playground"}
			// env, _ := pterm.DefaultInteractiveSelect.
			// 	WithDefaultText("select the environment you want to initialize for").
			// 	WithOptions(envs).
			// 	Show()
			// hd := consts.Hubs[env]
			//
			// var raID string
			// if len(args) != 0 {
			// 	raID = args[0]
			// } else {
			// 	raID, _ = pterm.DefaultInteractiveTextInput.WithDefaultText(
			// 		"provide RollApp ID you plan to run the nodes for",
			// 	).Show()
			// }
			//
			// dymdBinaryOptions := types.Dependency{
			// 	Name:       "dymension",
			// 	Repository: "https://github.com/artemijspavlovs/dymension",
			// 	Release:    "v3.1.0-pg07",
			// 	Binaries: []types.BinaryPathPair{
			// 		{
			// 			Binary:            "dymd",
			// 			BinaryDestination: consts.Executables.Dymension,
			// 			BuildCommand: exec.Command(
			// 				"make",
			// 				"build",
			// 			),
			// 		},
			// 	},
			// }
			// pterm.Info.Println("installing dependencies")
			// err = dependencies.InstallBinaryFromRelease(dymdBinaryOptions)
			// if err != nil {
			// 	pterm.Error.Println("failed to install dymd: ", err)
			// 	return
			// }
			//
			// raID = strings.TrimSpace(raID)
			//
			// getRaCmd := rollapp.GetRollappCmd(raID, hd)
			// var raResponse rollapp.ShowRollappResponse
			// out, err := bash.ExecCommandWithStdout(getRaCmd)
			// if err != nil {
			// 	pterm.Error.Println("failed to get rollapp: ", err)
			// 	return
			// }
			//
			// err = json.Unmarshal(out.Bytes(), &raResponse)
			// if err != nil {
			// 	pterm.Error.Println("failed to unmarshal", err)
			// 	return
			// }
			//
			// pterm.Info.Println("installing dependencies")
			// start := time.Now()
			// if raResponse.Rollapp.GenesisInfo.Bech32Prefix == "" {
			// 	pterm.Error.Println("no bech")
			// 	return
			// }
			// err = dependencies.InstallBinaries(
			// 	raResponse.Rollapp.GenesisInfo.Bech32Prefix,
			// 	false,
			// 	strings.ToLower(raResponse.Rollapp.VmType),
			// 	rollerData,
			// )
			// if err != nil {
			// 	pterm.Error.Println("failed to install binaries: ", err)
			// 	return
			// }
			// elapsed := time.Since(start)
			// fmt.Println("time elapsed: ", elapsed)
		},
	}

	cmd.Flags().String("node", consts.PlaygroundHubData.RpcUrl, "hub rpc endpoint")
	cmd.Flags().String("chain-id", consts.PlaygroundHubData.ID, "hub chain id")

	return cmd
}
