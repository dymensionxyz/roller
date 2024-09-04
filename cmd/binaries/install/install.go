package install

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/rollapp"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
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
					"provide a rollapp ID that you want to run the node for",
				).Show()
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

			installBinaries(raResponse.Rollapp.Bech32Prefix)
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
					BuildCommand:      exec.Command("make", "build"),
					BuildArgs: []string{
						bech32,
					},
				},
			},
		},
	}

	for k, dep := range deps {
		{
			err := cloneAndBuild(dep, k)
			if err != nil {
				pterm.Error.Printf("failed to build binary %s: %v", k, err)
				return
			}
		}
	}
}

func cloneAndBuild(dep Dependency, td string) error {
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
	fmt.Println(c.String())
	if err := c.Run(); err != nil {
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
			log.Fatalf("Failed to checkout version for %s: %v", dep.Repository, err)
		}
	} else {
		spinner.UpdateText("starting build from main branch")
	}

	// Build the binary
	for _, binary := range dep.Binaries {
		c := exec.Command(binary.BuildCommand.String(), binary.BuildArgs...)
		if err := c.Run(); err != nil {
			spinner.Fail("failed to build binary %s: %v", binary.BuildCommand.String())
			return err
		}

		if err := os.Rename(binary.BuildDestination, binary.BinaryDestination); err != nil {
			spinner.Fail(
				fmt.Sprintf(
					"Failed to move binary %s to %s",
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