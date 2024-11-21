package eibc

import (
	"time"

	"github.com/ignite/cli/ignite/pkg/cosmosaccount"
)

type Config struct {
	Fulfillers fulfillerConfig `yaml:"fulfillers"`
	Gas        GasConfig       `yaml:"gas"`

	LogLevel    string `yaml:"log_level"`
	NodeAddress string `yaml:"node_address"`

	OperatorConfig operatorConfig           `yaml:"operator"`
	OrderPolling   orderPollingConfig       `yaml:"order_polling"`
	Rollapps       map[string]rollappConfig `yaml:"rollapps"`

	SlackConfig slackConfig      `yaml:"slack"`
	Validation  validationConfig `yaml:"validation"`
}

type orderPollingConfig struct {
	IndexerURL string        `yaml:"indexer_url"`
	Interval   time.Duration `yaml:"interval"`
	Enabled    bool          `yaml:"enabled"`
}

type rollappConfig struct {
	FullNodes        []string `yaml:"full_nodes"`
	MinConfirmations string   `yaml:"min_confirmations"`
}

type GasConfig struct {
	Fees string `yaml:"fees"`
}

type fulfillerConfig struct {
	Scale          int                          `yaml:"scale"`
	KeyringBackend cosmosaccount.KeyringBackend `yaml:"keyring_backend"`
	KeyringDir     string                       `yaml:"keyring_dir"`
	MaxOrdersPerTx int                          `yaml:"max_orders_per_tx"`
	PolicyAddress  string                       `yaml:"policy_address"`
}

type operatorConfig struct {
	AccountName    string                       `yaml:"account_name"`
	GroupID        string                       `yaml:"group_id"`
	KeyringBackend cosmosaccount.KeyringBackend `yaml:"keyring_backend"`
	KeyringDir     string                       `yaml:"keyring_dir"`
	MinFeeShare    float32                      `yaml:"min_fee_share"`
}

type validationConfig struct {
	FallbackLevel string `yaml:"fallback_level"`
	WaitTime      string `yaml:"wait_time"`
	Interval      string `yaml:"interval"`
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
	delete(e.Rollapps, chainId)
}
