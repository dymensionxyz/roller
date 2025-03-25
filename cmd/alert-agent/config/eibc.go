package config

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v3"

	initam "github.com/dymensionxyz/roller/cmd/alert-agent/init"
	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	eibcutils "github.com/dymensionxyz/roller/utils/eibc"
	"github.com/dymensionxyz/roller/utils/roller"
)

func AddEibcLPCmd() *cobra.Command {
	var rollappList string

	cmd := &cobra.Command{
		Use:  "add-eibc-lp <grantee-address> <threshold-amount>",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			granteeAddr := args[0]
			amount := args[1]
			home := cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String()

			rollerData, err := roller.LoadConfig(home)
			if err != nil {
				return fmt.Errorf("failed to load roller config: %w", err)
			}

			configDir := filepath.Join(home, consts.ConfigDirName.AlertAgent)
			if _, err := os.Stat(configDir); err != nil {
				return fmt.Errorf(
					"config directory not found, run `roller alert-agent init` first: %w",
					err,
				)
			}

			configPath := filepath.Join(configDir, "config.yaml")

			configByte, err := os.ReadFile(configPath)
			if err != nil {
				return fmt.Errorf("failed to read config file: %w", err)
			}

			var config initam.AlertConfig
			err = yaml.Unmarshal(configByte, config)
			if err != nil {
				return fmt.Errorf("failed to unmarshal config: %w", err)
			}

			grantsByGrantee, err := eibcutils.GetGrantsByGrantee(granteeAddr, rollerData.HubData)
			if err != nil {
				return fmt.Errorf("failed to get grants-by-gratee: %w", err)
			}

			var selectedRollapps []string
			if rollappList != "" {
				selectedRollapps = strings.Split(rollappList, ",")
				for _, grant := range grantsByGrantee.Grants {
					for _, rollapp := range grant.Authorization.Value.Rollapps {
						if !slices.Contains(selectedRollapps, rollapp.RollappID) {
							continue
						}
						for _, spendLimit := range rollapp.SpendLimit {
							for _, addr := range config.Addresses {
								if addr.Address == grant.Granter && addr.Threshold.Denom == spendLimit.Denom {
									continue
								}
							}
							config.Addresses = append(config.Addresses, initam.AddressConfig{
								Name: fmt.Sprintf(
									"%s LP wallet",
									strings.ToUpper(rollerData.RollappID),
								),
								RestEndpoint: rollerData.HubData.ApiUrl,
								Address:      grant.Granter,
								Threshold: initam.AddressThreshold{
									Denom:  spendLimit.Denom,
									Amount: amount,
								},
							})
						}
					}
				}
			} else {
				for _, grant := range grantsByGrantee.Grants {
					for _, rollapp := range grant.Authorization.Value.Rollapps {
						for _, spendLimit := range rollapp.SpendLimit {
							for _, addr := range config.Addresses {
								if addr.Address == grant.Granter && addr.Threshold.Denom == spendLimit.Denom {
									continue
								}
							}
							config.Addresses = append(config.Addresses, initam.AddressConfig{
								Name: fmt.Sprintf(
									"%s LP wallet",
									strings.ToUpper(rollerData.RollappID),
								),
								RestEndpoint: rollerData.HubData.ApiUrl,
								Address:      grant.Granter,
								Threshold: initam.AddressThreshold{
									Denom:  spendLimit.Denom,
									Amount: amount,
								},
							})
						}
					}
				}
			}

			updatedConfig, err := yaml.Marshal(&config)
			if err != nil {
				return fmt.Errorf("failed to marshal updated config: %w", err)
			}

			err = os.WriteFile(configPath, updatedConfig, 0644)
			if err != nil {
				return fmt.Errorf("failed to write config file: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&rollappList, "rollapp", "", "Comma-separated list of rollapps exp: --rollapps a,b,c (default: all)")
	return cmd
}

func RemoveEibcLPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "remove-eibc-lp <grantee-address>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			granteeAddr := args[0]
			home := cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String()

			rollerData, err := roller.LoadConfig(home)
			if err != nil {
				return fmt.Errorf("failed to load roller config: %w", err)
			}

			configDir := filepath.Join(home, consts.ConfigDirName.AlertAgent)
			if _, err := os.Stat(configDir); err != nil {
				return fmt.Errorf(
					"config directory not found, run `roller alert-agent init` first: %w",
					err,
				)
			}

			configPath := filepath.Join(configDir, "config.yaml")

			configByte, err := os.ReadFile(configPath)
			if err != nil {
				return fmt.Errorf("failed to read config file: %w", err)
			}

			var config initam.AlertConfig
			err = yaml.Unmarshal(configByte, config)
			if err != nil {
				return fmt.Errorf("failed to marshal config: %w", err)
			}

			grantsByGrantee, err := eibcutils.GetGrantsByGrantee(granteeAddr, rollerData.HubData)
			if err != nil {
				return fmt.Errorf("failed to get grants-by-gratee: %w", err)
			}
			for _, grant := range grantsByGrantee.Grants {
				for i, addr := range config.Addresses {
					if addr.Address == grant.Granter {
						config.Addresses = append(config.Addresses[:i], config.Addresses[i+1:]...)
					}
				}
			}

			updatedConfig, err := yaml.Marshal(&config)
			if err != nil {
				return fmt.Errorf("failed to marshal updated config: %w", err)
			}

			err = os.WriteFile(configPath, updatedConfig, 0644)
			if err != nil {
				return fmt.Errorf("failed to write config file: %w", err)
			}

			return nil
		},
	}

	return cmd
}
