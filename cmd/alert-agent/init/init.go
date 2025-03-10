package initam

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v3"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/dependencies"
	eibcutils "github.com/dymensionxyz/roller/utils/eibc"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/roller"
)

type MetricConfig struct {
	Name         string `yaml:"name"`
	RestEndpoint string `yaml:"rest_endpoint"`
	Metric       string `yaml:"metric"`
	Threshold    int64  `yaml:"threshold"`
}

type AddressThreshold struct {
	Denom  string `yaml:"denom"`
	Amount string `yaml:"amount"`
}

type AddressConfig struct {
	Name         string           `yaml:"name"`
	RestEndpoint string           `yaml:"rest_endpoint"`
	Address      string           `yaml:"address"`
	Threshold    AddressThreshold `yaml:"threshold"`
}

type TelegramConfig struct {
	BotToken string `yaml:"bot_token"`
	ChatID   string `yaml:"chat_id"`
}

type AlertConfig struct {
	CheckInterval int64           `yaml:"check_interval"`
	AlertCooldown int64           `yaml:"alert_cooldown"`
	Metrics       []MetricConfig  `yaml:"metrics"`
	Addresses     []AddressConfig `yaml:"addresses"`
	Telegram      TelegramConfig  `yaml:"telegram"`
}

func getDefaultMetrics() []MetricConfig {
	return []MetricConfig{
		{
			Name:         "Mempool Size",
			RestEndpoint: "http://localhost:2112/metrics",
			Metric:       "dymint_mempool_size",
			Threshold:    20,
		},
		{
			Name:         "Pending Submissions Skew",
			RestEndpoint: "http://localhost:2112/metrics",
			Metric:       "rollapp_pending_submissions_skew_batches",
			Threshold:    30,
		},
		{
			Name:         "Failed DA Submissions",
			RestEndpoint: "http://localhost:2112/metrics",
			Metric:       "rollapp_consecutive_failed_da_submissions",
			Threshold:    30,
		},
	}
}

func getDefaultConfig() AlertConfig {
	return AlertConfig{
		CheckInterval: 600,
		AlertCooldown: 3600,
		Metrics:       getDefaultMetrics(),
		Addresses:     []AddressConfig{},
		Telegram: TelegramConfig{
			BotToken: "<your-bot-token>",
			ChatID:   "<your-chat-id>",
		},
	}
}

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize the Alert Agent configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			home := cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String()

			configDir := filepath.Join(home, consts.ConfigDirName.AlertAgent)
			if err := os.MkdirAll(configDir, 0o755); err != nil {
				return fmt.Errorf("failed to create config directory: %w", err)
			}

			err := os.MkdirAll(configDir, 0o755)
			if err != nil {
				return fmt.Errorf("failed to create config directory %s: %w", configDir, err)
			}

			configPath := filepath.Join(configDir, "config.yaml")
			if _, err := os.Stat(configPath); err == nil {
				return fmt.Errorf("config file already exists at %s", configPath)
			}

			_, err = os.Create(configPath)
			if err != nil {
				return fmt.Errorf("failed to create config file: %w", err)
			}

			config := getDefaultConfig()

			rollerData, err := roller.LoadConfig(home)
			if err != nil {
				return fmt.Errorf("failed to load roller config: %w", err)
			}

			dep := dependencies.DefaultAlertAgentDependency()
			if err := dependencies.InstallBinaryFromRelease(dep); err != nil {
				return fmt.Errorf("failed to install alert manager: %w", err)
			}

			keyInfos, err := keys.All(rollerData, rollerData.HubData)
			if err != nil {
				return fmt.Errorf("failed to get keys: %w", err)
			}

			thresholdAmounts := map[string]string{
				consts.KeysIds.HubSequencer: "10000000000000000000", // 10 DYM
				consts.KeysIds.HubRelayer:   "2000000000000000000",  // 2 DYM
				consts.KeysIds.Eibc:         "10000000000000000000", // 10 DYM
			}

			for _, keyInfo := range keyInfos {
				amount, ok := thresholdAmounts[keyInfo.Name]
				if !ok {
					amount = "2000000000000000000" // 2 DYM default
				}

				config.Addresses = append(config.Addresses, AddressConfig{
					Name: fmt.Sprintf(
						"%s %s wallet",
						strings.ToUpper(rollerData.RollappID),
						keyInfo.Name,
					),
					RestEndpoint: rollerData.HubData.ApiUrl,
					Address:      keyInfo.Address,
					Threshold: AddressThreshold{
						Denom:  "adym",
						Amount: amount,
					},
				})

				if keyInfo.Name == consts.KeysIds.Eibc {
					eibcHome := filepath.Join(home, consts.ConfigDirName.Eibc)
					eibcConfigPath := filepath.Join(eibcHome, "config.yaml")
					eibcConfig, err := eibcutils.ReadConfig(eibcConfigPath)
					if err != nil {
						pterm.Error.Println("failed to read eibc config", err)
					} else {
						grantsByGrantee, err := eibcutils.GetGrantsByGrantee(eibcConfig.Fulfillers.PolicyAddress)
						if err != nil {
							return fmt.Errorf("failed to get grants-by-gratee: %w", err)
						}
						for _, grant := range grantsByGrantee.Grants {
							for _, rollapp := range grant.Authorization.Value.Rollapps {
								for _, spendLimit := range rollapp.SpendLimit {
									config.Addresses = append(config.Addresses, AddressConfig{
										Name: fmt.Sprintf(
											"%s LP wallet",
											strings.ToUpper(rollerData.RollappID),
										),
										RestEndpoint: rollerData.HubData.ApiUrl,
										Address:      grant.Granter,
										Threshold: AddressThreshold{
											Denom:  spendLimit.Denom,
											Amount: amount,
										},
									})
								}
							}
						}
					}
				}
			}

			tgToken, _ := pterm.DefaultInteractiveTextInput.WithDefaultText(
				"Please provide a telegram bot token, learn how to obtain one here: https://github.com/dymensionxyz/alert-agent?tab=readme-ov-file#telegram-setup-optional",
			).Show()
			tgChatId, _ := pterm.DefaultInteractiveTextInput.WithDefaultText(
				"Please provide a telegram chat id, learn how to obtain it for group or telegram handle here: https://github.com/dymensionxyz/alert-agent?tab=readme-ov-file#telegram-setup-optional",
			).Show()

			config.Telegram.BotToken = tgToken
			config.Telegram.ChatID = tgChatId

			yamlData, err := yaml.Marshal(config)
			if err != nil {
				return fmt.Errorf("failed to marshal config: %w", err)
			}

			if err := os.WriteFile(configPath, yamlData, 0o644); err != nil {
				return fmt.Errorf("failed to write config file: %w", err)
			}

			fmt.Printf("Created alert agent config at %s\n", configPath)
			pterm.Warning.Println(
				"By default, only dymension addresses were added to the config, please add the DA address yourself",
			)
			return nil
		},
	}

	initconfig.AddGlobalFlags(cmd)
	return cmd
}
