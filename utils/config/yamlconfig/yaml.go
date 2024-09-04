package yamlconfig

import (
	"fmt"
	"time"

	"github.com/ignite/cli/ignite/pkg/cosmosaccount"
	"gopkg.in/yaml.v3"
)

func UpdateNestedYAML(data map[interface{}]interface{}, keyPath []string, value interface{}) error {
	if len(keyPath) == 0 {
		return fmt.Errorf("empty key path")
	}
	if len(keyPath) == 1 {
		if value == nil {
			delete(data, keyPath[0])
		} else {
			data[keyPath[0]] = value
		}
		return nil
	}
	nextMap, ok := data[keyPath[0]].(map[interface{}]interface{})
	if !ok {
		return fmt.Errorf("failed to set nested map for key: %s", keyPath[0])
	}
	return UpdateNestedYAML(nextMap, keyPath[1:], value)
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
	LogLevel string `mapstructure:"log_level"`
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
