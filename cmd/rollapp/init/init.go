package initrollapp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/archives"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/dependencies"
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
			isMockFlagSet := cmd.Flags().Changed("mock")
			shouldUseMockBackend, _ := cmd.Flags().GetBool("mock")

			var raID string
			if len(args) != 0 {
				raID = args[0]
			} else {
				raID, _ = pterm.DefaultInteractiveTextInput.WithDefaultText(
					"provide a rollapp ID that you want to run the node for",
				).Show()
			}

			var hd consts.HubData
			var env string

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
			err = installBinaryFromRelease(dymdBinaryOptions)
			if err != nil {
				pterm.Error.Println("failed to install dymd: ", err)
				return
			}

			if shouldUseMockBackend {
				env := "mock"
				err = installBinaries("mock")
				if err != nil {
					pterm.Error.Println("failed to install binaries: ", err)
					return
				}
				err := runInit(cmd, env, raID)
				if err != nil {
					fmt.Println("failed to run init: ", err)
					return
				}
				return
			}

			if !isMockFlagSet && !shouldUseMockBackend {
				envs := []string{"mock", "playground"}
				env, _ = pterm.DefaultInteractiveSelect.
					WithDefaultText("select the environment you want to initialize for").
					WithOptions(envs).
					Show()
				hd = consts.Hubs[env]
				if env == "mock" {
					err = installBinaries("mock")
					if err != nil {
						pterm.Error.Println("failed to install binaries: ", err)
						return
					}
					err := runInit(cmd, env, raID)
					if err != nil {
						fmt.Println("failed to run init: ", err)
						return
					}
					return
				}
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

			start := time.Now()
			if raResponse.Rollapp.GenesisInfo.Bech32Prefix == "" {
				pterm.Error.Println("no bech")
				return
			}
			err = installBinaries(raResponse.Rollapp.GenesisInfo.Bech32Prefix)
			if err != nil {
				pterm.Error.Println("failed to install binaries: ", err)
				return
			}
			elapsed := time.Since(start)
			fmt.Println("time elapsed: ", elapsed)
			// END: ex binaries install

			isRollappRegistered, _ := rollapp.IsRollappRegistered(raID, hd)
			if !isRollappRegistered {
				pterm.Error.Printf("%s was not found as a registered rollapp: %v", raID, err)
				return
			}

			err = json.Unmarshal(out.Bytes(), &raResponse)
			if err != nil {
				pterm.Error.Println("failed to unmarshal", err)
				return
			}

			bp, err := rollapp.ExtractBech32Prefix()
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

			err = runInit(cmd, env, raID)
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

func installBinaries(bech32 string) error {
	c := exec.Command("sudo", "mkdir", "-p", consts.InternalBinsDir)
	_, err := bash.ExecCommandWithStdout(c)
	if err != nil {
		errMsg := fmt.Sprintf("failed to create %s", consts.InternalBinsDir)
		return errors.New(errMsg)
	}

	defer func() {
		dir, err := os.UserHomeDir()
		if err != nil {
			return
		}
		_ = os.Chdir(dir)
	}()

	buildableDeps := map[string]dependencies.Dependency{
		"rollapp": {
			Repository: "https://github.com/dymensionxyz/rollapp-evm.git",
			Release:    "e68f8190f1301b317846623a9e83be7acc2ad56e", // 20240909 rolapparams module
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
			errMsg := fmt.Sprintf("failed to build binary %s: %v", k, err)
			return errors.New(errMsg)
		}

	}

	for k, dep := range buildableDeps {
		err := installBinaryFromRepo(dep, k)
		if err != nil {
			errMsg := fmt.Sprintf("failed to build binary %s: %v", k, err)
			return errors.New(errMsg)
		}
	}

	return nil
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
	if goArch == "amd64" && dep.Name == "celestia-app" {
		goArch = "x86_64"
	}

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

	// nolint errcheck
	defer resp.Body.Close()
	err = archives.ExtractTarGz(destination, resp.Body, dep)
	if err != nil {
		return err
	}

	return nil
}
