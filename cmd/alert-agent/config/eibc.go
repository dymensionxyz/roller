package config

import (
	"fmt"
	"os"
	"path/filepath"
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
				return fmt.Errorf("failed to marshal config: %w", err)
			}

			grantsByGrantee, err := eibcutils.GetGrantsByGrantee(granteeAddr)
			if err != nil {
				return fmt.Errorf("failed to get grants-by-gratee: %w", err)
			}
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

			return nil
		},
	}

	return cmd
}

func RemoveEibcLPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "remove-eibc-lp <grantee-address>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			granteeAddr := args[0]
			home := cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String()

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

			grantsByGrantee, err := eibcutils.GetGrantsByGrantee(granteeAddr)
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

			return nil
		},
	}

	return cmd
}
