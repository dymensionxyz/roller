package yamlconfig

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ignite/cli/ignite/pkg/cosmosaccount"
	"gopkg.in/yaml.v3"
)

func UpdateNestedYAML(filename string, updates map[string]interface{}) error {
	// Read YAML file
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	// Parse YAML
	var yamlData map[string]interface{}
	err = yaml.Unmarshal(data, &yamlData)
	if err != nil {
		return err
	}

	// Update values
	for path, value := range updates {
		keys := strings.Split(path, ".")
		err = setNestedValue(yamlData, keys, value)
		if err != nil {
			return fmt.Errorf("error updating %s: %v", path, err)
		}
	}

	// Marshal back to YAML
	updatedData, err := yaml.Marshal(yamlData)
	if err != nil {
		return err
	}

	// Write updated YAML back to file
	return os.WriteFile(filename, updatedData, 0o644)
}

func setNestedValue(data map[string]interface{}, keys []string, value interface{}) error {
	for i, key := range keys {
		if i == len(keys)-1 {
			data[key] = value
			return nil
		}

		if _, ok := data[key]; !ok {
			data[key] = make(map[string]interface{})
		}

		nestedMap, ok := data[key].(map[string]interface{})
		if !ok {
			return fmt.Errorf("failed to set nested map for key: %s", key)
		}

		data = nestedMap
	}
	return nil
}

type EibcConfig struct {
	HomeDir      string             `mapstructure:"home_dir"`
	NodeAddress  string             `mapstructure:"node_address"`
	DBPath       string             `mapstructure:"db_path"`
	Gas          GasConfig          `mapstructure:"gas"`
	OrderPolling OrderPollingConfig `mapstructure:"order_polling"`

	Whale           whaleConfig     `mapstructure:"whale"`
	Bots            botConfig       `mapstructure:"bots"`
	FulfillCriteria fulfillCriteria `mapstructure:"fulfill_criteria"`
	LogLevel        string          `mapstructure:"log_level"`
}

type OrderPollingConfig struct {
	IndexerURL string        `mapstructure:"indexer_url"`
	Interval   time.Duration `mapstructure:"interval"`
	Enabled    bool          `mapstructure:"enabled"`
}

type GasConfig struct {
	Prices            string `mapstructure:"prices"`
	Fees              string `mapstructure:"fees"`
	MinimumGasBalance string `mapstructure:"minimum_gas_balance"`
}

type botConfig struct {
	NumberOfBots   int                          `mapstructure:"number_of_bots"`
	KeyringBackend cosmosaccount.KeyringBackend `mapstructure:"keyring_backend"`
	KeyringDir     string                       `mapstructure:"keyring_dir"`
	TopUpFactor    int                          `mapstructure:"top_up_factor"`
	MaxOrdersPerTx int                          `mapstructure:"max_orders_per_tx"`
}

type whaleConfig struct {
	AccountName              string                       `mapstructure:"account_name"`
	KeyringBackend           cosmosaccount.KeyringBackend `mapstructure:"keyring_backend"`
	KeyringDir               string                       `mapstructure:"keyring_dir"`
	AllowedBalanceThresholds map[string]string            `mapstructure:"allowed_balance_thresholds"`
}

type fulfillCriteria struct {
	MinFeePercentage minFeePercentage `mapstructure:"min_fee_percentage"`
}

type minFeePercentage struct {
	Chain map[string]float32 `mapstructure:"chain"`
	Asset map[string]float32 `mapstructure:"asset"`
}

type slackConfig struct {
	Enabled   bool   `mapstructure:"enabled"`
	BotToken  string `mapstructure:"bot_token"`
	AppToken  string `mapstructure:"app_token"`
	ChannelID string `mapstructure:"channel_id"`
}

func (e *EibcConfig) RemoveChain(chainId string) {
	delete(e.FulfillCriteria.MinFeePercentage.Chain, chainId)
}

func (e *EibcConfig) RemoveDenom(denom string) {
	delete(e.FulfillCriteria.MinFeePercentage.Asset, denom)
}
