package eibc

import (
	"time"

	"github.com/ignite/cli/ignite/pkg/cosmosaccount"
)

type Config struct {
	HomeDir      string             `yaml:"home_dir"`
	NodeAddress  string             `yaml:"node_address"`
	DBPath       string             `yaml:"db_path"`
	Gas          GasConfig          `yaml:"gas"`
	OrderPolling OrderPollingConfig `yaml:"order_polling"`

	Whale           whaleConfig     `yaml:"whale"`
	Bots            botConfig       `yaml:"bots"`
	FulfillCriteria fulfillCriteria `yaml:"fulfill_criteria"`

	LogLevel    string      `yaml:"log_level"`
	SlackConfig slackConfig `yaml:"slack"`
	SkipRefund  bool        `yaml:"skip_refund"`
}

type OrderPollingConfig struct {
	IndexerURL string        `yaml:"indexer_url"`
	Interval   time.Duration `yaml:"interval"`
	Enabled    bool          `yaml:"enabled"`
}

type GasConfig struct {
	Prices            string `yaml:"prices"`
	Fees              string `yaml:"fees"`
	MinimumGasBalance string `yaml:"minimum_gas_balance"`
}

type botConfig struct {
	NumberOfBots   int                          `yaml:"number_of_bots"`
	KeyringBackend cosmosaccount.KeyringBackend `yaml:"keyring_backend"`
	KeyringDir     string                       `yaml:"keyring_dir"`
	TopUpFactor    int                          `yaml:"top_up_factor"`
	MaxOrdersPerTx int                          `yaml:"max_orders_per_tx"`
}

type whaleConfig struct {
	AccountName              string                       `yaml:"account_name"`
	KeyringBackend           cosmosaccount.KeyringBackend `yaml:"keyring_backend"`
	KeyringDir               string                       `yaml:"keyring_dir"`
	AllowedBalanceThresholds map[string]string            `yaml:"allowed_balance_thresholds"`
}

type fulfillCriteria struct {
	MinFeePercentage minFeePercentage `yaml:"min_fee_percentage"`
}

type minFeePercentage struct {
	Chain map[string]float32 `yaml:"chain"`
	Asset map[string]float32 `yaml:"asset"`
}

type slackConfig struct {
	Enabled   bool   `yaml:"enabled"`
	BotToken  string `yaml:"bot_token"`
	AppToken  string `yaml:"app_token"`
	ChannelID string `yaml:"channel_id"`
}

func (e *Config) RemoveChain(chainId string) {
	delete(e.FulfillCriteria.MinFeePercentage.Chain, chainId)
}

func (e *Config) RemoveAllowedBalanceThreshold(denom string) {
	delete(e.Whale.AllowedBalanceThresholds, denom)
}

func (e *Config) RemoveDenom(denom string) {
	delete(e.FulfillCriteria.MinFeePercentage.Asset, denom)
}
