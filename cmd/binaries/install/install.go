package install

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/rollapp"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install <rollapp-id>",
		Short: "Send the DYM rewards associated with the given private key to the destination address",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var raID string
			if len(args) != 0 {
				raID = args[0]
			} else {
				raID, _ = pterm.DefaultInteractiveTextInput.WithDefaultText(
					"provide RollApp ID you plan to run the nodes for",
				).Show()
			}

			dymdBinaryOptions := Dependency{
				Repository: "https://github.com/dymensionxyz/dymension.git",
				Commit:     "playground/v1-rc03",
				Binaries: []BinaryPathPair{
					{
						BuildDestination:  "./build/dymd",
						BinaryDestination: consts.Executables.Dymension,
						BuildCommand: exec.Command(
							"make",
							"build",
						),
					},
				},
			}

			err := installBinaryFromRepo(dymdBinaryOptions, "dymd")
			if err != nil {
				pterm.Error.Println("failed to install dymd", err)
				return
			}

			envs := []string{"devnet", "playground"}
			env, _ := pterm.DefaultInteractiveSelect.
				WithDefaultText("select the environment you want to initialize for").
				WithOptions(envs).
				Show()
			hd := consts.Hubs[env]

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

			start := time.Now()
			installBinaries(raResponse.Rollapp.Bech32Prefix)
			elapsed := time.Since(start)
			fmt.Println("time elapsed: ", elapsed)
		},
	}

	cmd.Flags().String("node", consts.PlaygroundHubData.RPC_URL, "hub rpc endpoint")
	cmd.Flags().String("chain-id", consts.PlaygroundHubData.ID, "hub chain id")

	return cmd
}

type BinaryPathPair struct {
	BuildDestination  string
	BinaryDestination string
	BuildCommand      *exec.Cmd
	BuildArgs         []string
}

type Dependency struct {
	Repository string
	Commit     string
	Binaries   []BinaryPathPair
}

func installBinaries(bech32 string) {
	deps := map[string]Dependency{
		"rollapp": {
			Repository: "https://github.com/dymensionxyz/rollapp-evm.git",
			Commit:     "main",
			Binaries: []BinaryPathPair{
				{
					BuildDestination:  "./build/rollapp-evm",
					BinaryDestination: consts.Executables.RollappEVM,
					BuildCommand: exec.Command(
						"make",
						"build",
						fmt.Sprintf("BECH32_PREFIX=%s", bech32),
					),
				},
			},
		},
		"roller": {
			Repository: "https://github.com/dymensionxyz/roller.git",
			Commit:     "main",
			Binaries: []BinaryPathPair{
				{
					BuildDestination:  "./build/roller",
					BinaryDestination: consts.Executables.Roller,
					BuildCommand: exec.Command(
						"make",
						"build",
					),
				},
			},
		},
		// "celestia": {
		// 	Repository: "https://github.com/celestiaorg/celestia-node.git",
		// 	Commit:     "v0.16.0-rc0",
		// 	Binaries: []BinaryPathPair{
		// 		{
		// 			BuildDestination:  "./build/celestia",
		// 			BinaryDestination: consts.Executables.Celestia,
		// 			BuildCommand: exec.Command(
		// 				"make",
		// 				"build",
		// 			),
		// 		},
		// 		{
		// 			BuildDestination:  "./cel-key",
		// 			BinaryDestination: consts.Executables.CelKey,
		// 			BuildCommand: exec.Command(
		// 				"make",
		// 				"cel-key",
		// 			),
		// 		},
		// 	},
		// },
		// "celestia-app": {
		// 	Repository: "https://github.com/celestiaorg/celestia-app.git",
		// 	Commit:     "v2.0.0",
		// 	Binaries: []BinaryPathPair{
		// 		{
		// 			BuildDestination:  "./build/celestia-appd",
		// 			BinaryDestination: consts.Executables.CelestiaApp,
		// 			BuildCommand: exec.Command(
		// 				"make",
		// 				"build",
		// 			),
		// 		},
		// 	},
		// },
		"eibc-client": {
			Repository: "https://github.com/dymensionxyz/eibc-client.git",
			Commit:     "main",
			Binaries: []BinaryPathPair{
				{
					BuildDestination:  "./build/eibc-client",
					BinaryDestination: consts.Executables.Eibc,
					BuildCommand: exec.Command(
						"make",
						"build",
					),
				},
			},
		},
		"rly": {
			Repository: "https://github.com/dymensionxyz/go-relayer.git",
			Commit:     "v0.3.4-v2.5.2-relayer-canon-1",
			Binaries: []BinaryPathPair{
				{
					BuildDestination:  "./build/rly",
					BinaryDestination: consts.Executables.Relayer,
					BuildCommand: exec.Command(
						"make",
						"build",
					),
				},
			},
		},
	}

	for k, dep := range deps {
		{
			err := installBinaryFromRepo(dep, k)
			if err != nil {
				pterm.Error.Printf("failed to build binary %s: %v", k, err)
				return
			}
		}
	}
}

func installBinaryFromRepo(dep Dependency, td string) error {
	targetDir, err := os.MkdirTemp(os.TempDir(), td)
	if err != nil {
		return err
	}
	// defer os.RemoveAll(targetDir)
	// Clone the repository
	err = os.Chdir(targetDir)
	if err != nil {
		pterm.Error.Println("failed to create a temp directory")
		return err
	}

	spinner, _ := pterm.DefaultSpinner.Start(
		fmt.Sprintf("cloning %s into %s", dep.Repository, targetDir),
	)

	c := exec.Command("git", "clone", dep.Repository, targetDir)

	_, err = bash.ExecCommandWithStdout(c)
	if err != nil {
		pterm.Error.Println("failed to clone")
		return err
	}
	// Change directory to the cloned repo
	if err := os.Chdir(targetDir); err != nil {
		pterm.Error.Println("failed to create a temp directory")
		return err
	}

	if dep.Commit != "main" {
		// Checkout a specific version (e.g., a tag or branch)
		spinner.UpdateText(fmt.Sprintf("checking out %s", dep.Commit))
		if err := exec.Command("git", "checkout", dep.Commit).Run(); err != nil {
			spinner.Fail(fmt.Sprintf("failed to checkout: %v\n", err))
			return err
		}
	}

	spinner.UpdateText(
		fmt.Sprintf(
			"starting build from %s (this can take several minutes)",
			dep.Commit,
		),
	)

	// Build the binary
	for _, binary := range dep.Binaries {
		_, err := bash.ExecCommandWithStdout(binary.BuildCommand)
		spinner.UpdateText(fmt.Sprintf("building %s\n", binary.BuildDestination))
		if err != nil {
			spinner.Fail(fmt.Sprintf("failed to build binary %s: %v\n", binary.BuildCommand, err))
			return err
		}

		if err := os.Rename(binary.BuildDestination, binary.BinaryDestination); err != nil {
			spinner.Fail(
				fmt.Sprintf(
					"Failed to move binary %s to %s\n",
					binary.BuildDestination,
					binary.BinaryDestination,
				),
			)

			return err
		}
		spinner.Success(fmt.Sprintf("Successfully installed %s\n", binary.BinaryDestination))
	}

	return nil
}
