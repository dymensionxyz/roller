package start

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"slices"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/data_layer/aptos"
	"github.com/dymensionxyz/roller/data_layer/kaspa"
	"github.com/dymensionxyz/roller/data_layer/sui"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/migrations"
	"github.com/dymensionxyz/roller/utils/roller"
	sequencerutils "github.com/dymensionxyz/roller/utils/sequencer"
	servicemanager "github.com/dymensionxyz/roller/utils/service_manager"
	"github.com/dymensionxyz/roller/utils/upgrades"
)

func RollappCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the systemd services relevant to RollApp",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			home, err := filesystem.ExpandHomePath(
				cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String(),
			)
			if err != nil {
				pterm.Error.Println("failed to expand home directory")
				return
			}

			rollappConfig, err := roller.LoadConfig(home)
			errorhandling.PrettifyErrorIfExists(err)

			if rollappConfig.DA.Backend == consts.Sui {
				cfgPath := sui.GetCfgFilePath(home)
				suiConfig, err := sui.LoadConfigFromTOML(cfgPath)
				if err != nil {
					pterm.Error.Println("failed to load config", err)
					return
				}

				err = os.Setenv("SUI_MNEMONIC", suiConfig.Mnemonic)
				if err != nil {
					pterm.Error.Println("failed to set env", err)
					return
				}
			}

			if rollappConfig.DA.Backend == consts.Aptos {
				cfgPath := aptos.GetCfgFilePath(home)
				aptConfig, err := aptos.LoadConfigFromTOML(cfgPath)
				if err != nil {
					pterm.Error.Println("failed to load config", err)
					return
				}
				err = os.Setenv("APT_PRIVATE_KEY", aptConfig.PrivateKey)
				if err != nil {
					pterm.Error.Println("failed to set env", err)
					return
				}
			}

			if rollappConfig.DA.Backend == consts.Kaspa {
				err := ensureKaspaMnemonicCached(rollappConfig.Home)
				if err != nil {
					pterm.Error.Println("failed to prepare kaspa mnemonic:", err)
					return
				}
			}

			if rollappConfig.NodeType == "sequencer" {
				err = sequencerutils.CheckBalance(rollappConfig)
				if err != nil {
					pterm.Error.Println("failed to check sequencer balance: ", err)
					return
				}
			}

			if rollappConfig.HubData.ID != consts.MockHubID {
				raUpgrade, err := upgrades.NewRollappUpgrade(string(rollappConfig.RollappVMType))
				if err != nil {
					pterm.Error.Println("failed to check rollapp version equality: ", err)
				}

				err = migrations.RequireRollappMigrateIfNeeded(
					raUpgrade.CurrentVersionCommit[:6],
					rollappConfig.RollappBinaryVersion[:6],
					string(rollappConfig.RollappVMType),
				)
				if err != nil {
					pterm.Error.Println(err)
					return
				}
			}

			var servicesToStart []string
			if len(args) != 0 {
				if !slices.Contains(consts.RollappSystemdServices, args[0]) {
					pterm.Error.Printf(
						"invalid service name %s. Available services: %v\n",
						args[0],
						consts.RollappSystemdServices,
					)
					return
				}

				servicesToStart = []string{args[0]}
			} else {
				if rollappConfig.DA.Backend == consts.Celestia {
					servicesToStart = consts.RollappWithCelesSystemdServices
				} else {
					servicesToStart = consts.RollappSystemdServices
				}
			}

			if runtime.GOOS == "darwin" {
				err := startLaunchctlServices(servicesToStart)
				if err != nil {
					pterm.Error.Println("failed to start launchd services:", err)
					return
				}
			} else if runtime.GOOS == "linux" {
				err := startSystemdServices(servicesToStart)
				if err != nil {
					pterm.Error.Println("failed to start systemd services:", err)
					return
				}
			} else {
				pterm.Info.Printf(
					"the %s commands currently support only darwin and linux operating systems",
					cmd.Use,
				)
				return
			}

			defer func() {
				pterm.Info.Println("next steps:")
				pterm.Info.Printf(
					"run %s to set up IBC channels and start relaying packets\n",
					pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
						Sprintf("roller relayer setup"),
				)
				if runtime.GOOS == "linux" {
					pterm.Info.Printf(
						"run %s to view the logs  of the rollapp\n",
						pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
							Sprintf("journalctl -fu <service>"),
					)
				}
			}()
		},
	}
	return cmd
}

func RelayerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Starts the relayer locally",
		Run: func(cmd *cobra.Command, args []string) {
			if runtime.GOOS == "linux" {
				err := startSystemdServices(consts.RelayerSystemdServices)
				if err != nil {
					pterm.Error.Println("failed to start systemd services:", err)
					return
				}
			} else if runtime.GOOS == "darwin" {
				err := startLaunchctlServices(consts.RelayerSystemdServices)
				if err != nil {
					pterm.Error.Println("failed to start launchd services:", err)
					return
				}
			} else {
				pterm.Error.Printf(
					"the %s commands currently support only darwin and linux operating systems",
					cmd.Use,
				)
			}

			defer func() {
				pterm.Info.Println("next steps:")
				pterm.Info.Printf(
					"run %s to join the eibc market\n",
					pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
						Sprintf("roller eibc init"),
				)
				if runtime.GOOS == "linux" {
					pterm.Info.Printf(
						"run %s to view the logs of the relayer\n",
						pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
							Sprintf("journalctl -fu <service>"),
					)
				}
			}()
		},
	}
	return cmd
}

func EibcCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the eibc systemd services on local machine",
		Run: func(cmd *cobra.Command, args []string) {
			if runtime.GOOS == "linux" {
				err := startSystemdServices(consts.EibcSystemdServices)
				if err != nil {
					pterm.Error.Println("failed to start systemd services:", err)
					return
				}
			} else if runtime.GOOS == "darwin" {
				err := startLaunchctlServices(consts.EibcSystemdServices)
				if err != nil {
					pterm.Error.Println("failed to start launchd services:", err)
					return
				}
			} else {
				pterm.Error.Printf(
					"the %s commands currently support only darwin and linux operating systems",
					cmd.Use,
				)
			}

			defer func() {
				pterm.Info.Println("next steps:")
				pterm.Info.Println(
					"that's all folks",
				)

				if runtime.GOOS == "linux" {
					pterm.Info.Printf(
						"run %s to view the current status of the eibc client\n",
						pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
							Sprintf("journalctl -fu eibc"),
					)
				}
			}()
		},
	}
	return cmd
}

func OracleCmd(oracleType string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the systemd services on local machine",
		Run: func(cmd *cobra.Command, args []string) {
			switch oracleType {
			case "price":
				if runtime.GOOS == "linux" {
					err := startSystemdServices(consts.PriceOracleSystemdServices)
					if err != nil {
						pterm.Error.Println("failed to start systemd services:", err)
						return
					}
				} else if runtime.GOOS == "darwin" {
					err := startLaunchctlServices(consts.PriceOracleSystemdServices)
					if err != nil {
						pterm.Error.Println("failed to start launchd services:", err)
						return
					}
				} else {
					pterm.Error.Printf(
						"the %s commands currently support only darwin and linux operating systems",
						cmd.Use,
					)
				}
			case "rng":
				if runtime.GOOS == "linux" {
					err := startSystemdServices(consts.RngOracleSystemdServices)
					if err != nil {
						pterm.Error.Println("failed to start systemd services:", err)
						return
					}
				} else if runtime.GOOS == "darwin" {
					err := startLaunchctlServices(consts.RngOracleSystemdServices)
					if err != nil {
						pterm.Error.Println("failed to start launchd services:", err)
						return
					}
				} else {
					pterm.Error.Printf(
						"the %s commands currently support only darwin and linux operating systems",
						cmd.Use,
					)
				}
			default:
				pterm.Error.Println("invalid oracle type")
			}
		},
	}
	return cmd
}

func startSystemdServices(services []string) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf(
			"the services commands are only available on linux machines",
		)
	}
	for _, service := range services {
		err := servicemanager.StartSystemdService(
			fmt.Sprintf("%s.service", service),
		)
		if err != nil {
			return fmt.Errorf("failed to start %s systemd service: %v", service, err)
		}
	}
	pterm.Success.Printf(
		"💈 Services %s started successfully.\n",
		strings.Join(services, ", "),
	)
	return nil
}

func startLaunchctlServices(services []string) error {
	for _, service := range services {
		err := servicemanager.StartLaunchctlService(
			service,
		)
		if err != nil {
			return fmt.Errorf("failed to start %s launchctl service: %v", service, err)
		}
	}
	pterm.Success.Printf(
		"💈 Services %s started successfully.\n",
		strings.Join(services, ", "),
	)
	return nil
}

func ensureKaspaMnemonicCached(home string) error {
	cached, err := readKaspaMnemonicFromCache(home)
	if err != nil {
		return err
	}
	if cached != "" {
		return nil
	}

	mnemonic := strings.TrimSpace(os.Getenv(kaspa.MnemonicEnvVar))
	if mnemonic == "" {
		var promptErr error
		mnemonic, promptErr = pterm.DefaultInteractiveTextInput.WithDefaultText(
			"> Enter your Kaspa mnemonic",
		).Show()
		if promptErr != nil {
			return promptErr
		}
		mnemonic = strings.TrimSpace(mnemonic)
		if mnemonic == "" {
			return errors.New("kaspa mnemonic cannot be empty")
		}
	}

	if err := cacheKaspaMnemonic(home, mnemonic); err != nil {
		return err
	}
	return nil
}

func readKaspaMnemonicFromCache(home string) (string, error) {
	path := kaspa.GetMnemonicEnvFilePath(home)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", err
	}
	prefix := kaspa.MnemonicEnvVar + "="
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix)), nil
		}
	}
	return "", nil
}

func cacheKaspaMnemonic(home, mnemonic string) error {
	if mnemonic == "" {
		return errors.New("kaspa mnemonic cannot be empty")
	}
	envPath, err := kaspa.EnsureMnemonicEnvFile(home)
	if err != nil {
		return err
	}
	content := fmt.Sprintf("%s=%s\n", kaspa.MnemonicEnvVar, mnemonic)
	return os.WriteFile(envPath, []byte(content), 0o600)
}
