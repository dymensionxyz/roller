package install

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/archives"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/dependencies"
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
					"provide RollApp ID you plan to run the nodes for",
				).Show()
			}

			raID = strings.TrimSpace(raID)

			// TODO: instead of relying on dymd binary, query the rpc for rollapp
			dymdBinaryOptions := dependencies.Dependency{
				Name:       "dymension",
				Repository: "https://github.com/artemijspavlovs/dymension",
				Release:    "3.1.0-pg04",
				Binaries: []dependencies.BinaryPathPair{
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
			err := installBinaryFromRelease(dymdBinaryOptions)
			if err != nil {
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
			if raResponse.Rollapp.GenesisInfo.Bech32Prefix == "" {
				pterm.Error.Println("no bech")
				return
			}
			installBinaries(raResponse.Rollapp.GenesisInfo.Bech32Prefix)
			elapsed := time.Since(start)
			fmt.Println("time elapsed: ", elapsed)
		},
	}

	cmd.Flags().String("node", consts.PlaygroundHubData.RPC_URL, "hub rpc endpoint")
	cmd.Flags().String("chain-id", consts.PlaygroundHubData.ID, "hub chain id")

	return cmd
}

func installBinaries(bech32 string) {
	fmt.Println("bech:", bech32)
	buildableDeps := map[string]dependencies.Dependency{
		"rollapp": {
			Repository: "https://github.com/dymensionxyz/rollapp-evm.git",
			Release:    "559d878e83800717c885e89f2fbe619ee081b2a1", // 20240905 light client support
			Binaries: []dependencies.BinaryPathPair{
				{
					Binary:            "./build/rollapp-evm",
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
			Release:    "main",
			Binaries: []dependencies.BinaryPathPair{
				{
					Binary:            "./build/roller",
					BinaryDestination: consts.Executables.Roller,
					BuildCommand: exec.Command(
						"make",
						"build",
					),
				},
			},
		},
		"celestia": {
			Repository: "https://github.com/celestiaorg/celestia-node.git",
			Release:    "v0.16.0",
			Binaries: []dependencies.BinaryPathPair{
				{
					Binary:            "./build/celestia",
					BinaryDestination: consts.Executables.Celestia,
					BuildCommand: exec.Command(
						"make",
						"build",
					),
				},
				{
					Binary:            "./cel-key",
					BinaryDestination: consts.Executables.CelKey,
					BuildCommand: exec.Command(
						"make",
						"cel-key",
					),
				},
			},
		},
	}

	goreleaserDeps := map[string]dependencies.Dependency{
		"celestia-app": {
			Name:       "celestia-app",
			Repository: "https://github.com/celestiaorg/celestia-app",
			Release:    "2.1.2",
			Binaries: []dependencies.BinaryPathPair{
				{
					Binary:            "celestia-appd",
					BinaryDestination: consts.Executables.CelestiaApp,
					BuildCommand: exec.Command(
						"make",
						"build",
					),
				},
			},
		},
		"eibc-client": {
			Name:       "eibc-client",
			Repository: "https://github.com/artemijspavlovs/eibc-client",
			Release:    "1.1.0",
			Binaries: []dependencies.BinaryPathPair{
				{
					Binary:            "eibc-client",
					BinaryDestination: consts.Executables.Eibc,
				},
			},
		},
		"rly": {
			Name:       "go-relayer",
			Repository: "https://github.com/artemijspavlovs/go-relayer",
			Release:    "0.3.4-v2.5.2-relayer-canon-3",
			Binaries: []dependencies.BinaryPathPair{
				{
					Binary:            "rly",
					BinaryDestination: consts.Executables.Relayer,
				},
			},
		},
	}
	//
	for k, dep := range goreleaserDeps {
		err := installBinaryFromRelease(dep)
		if err != nil {
			pterm.Error.Printf("failed to build binary %s: %v", k, err)
			return
		}
	}

	for k, dep := range buildableDeps {
		err := installBinaryFromRepo(dep, k)
		if err != nil {
			pterm.Error.Printf("failed to build binary %s: %v", k, err)
			return
		}
	}
}

func installBinaryFromRepo(dep dependencies.Dependency, td string) error {
	targetDir, err := os.MkdirTemp(os.TempDir(), td)
	if err != nil {
		return err
	}
	// nolint: errcheck
	defer os.RemoveAll(targetDir)
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

	if dep.Release != "main" {
		// Checkout a specific version (e.g., a tag or branch)
		spinner.UpdateText(fmt.Sprintf("checking out %s", dep.Release))
		if err := exec.Command("git", "checkout", dep.Release).Run(); err != nil {
			spinner.Fail(fmt.Sprintf("failed to checkout: %v\n", err))
			return err
		}
	}

	spinner.UpdateText(
		fmt.Sprintf(
			"starting build from %s (this can take several minutes)",
			dep.Release,
		),
	)

	// Build the binary
	for _, binary := range dep.Binaries {
		_, err := bash.ExecCommandWithStdout(binary.BuildCommand)
		spinner.UpdateText(fmt.Sprintf("building %s\n", binary.Binary))
		if err != nil {
			spinner.Fail(fmt.Sprintf("failed to build binary %s: %v\n", binary.BuildCommand, err))
			return err
		}

		c := exec.Command("sudo", "mv", binary.Binary, binary.BinaryDestination)
		if _, err := bash.ExecCommandWithStdout(c); err != nil {
			spinner.Fail(
				fmt.Sprintf(
					"Failed to move binary %s to %s\n",
					binary.Binary,
					binary.BinaryDestination,
				),
			)
			return err
		}
		spinner.Success(fmt.Sprintf("Successfully installed %s\n", binary.BinaryDestination))
	}

	return nil
}

func installBinaryFromRelease(dep dependencies.Dependency) error {
	goOs := strings.Title(runtime.GOOS)
	goArch := strings.ToLower(runtime.GOARCH)
	targetDir, err := os.MkdirTemp(os.TempDir(), dep.Name)
	if err != nil {
		return err
	}
	archiveName := fmt.Sprintf(
		"%s_%s_%s.tar.gz",
		dep.Name,
		goOs,
		goArch,
	)
	// nolint: errcheck
	defer os.RemoveAll(targetDir)

	url := fmt.Sprintf(
		"%s/releases/download/v%s/%s",
		dep.Repository,
		dep.Release,
		archiveName,
	)

	err = downloadRelease(url, targetDir, dep)
	if err != nil {
		return err
	}

	return nil
}

func downloadRelease(url, destination string, dep dependencies.Dependency) error {
	// nolint gosec
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	err = archives.ExtractTarGz(destination, resp.Body, dep)
	if err != nil {
		return err
	}

	return nil
}
