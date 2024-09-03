package yamlconfig

import (
	"fmt"
	"time"

	"github.com/ignite/cli/ignite/pkg/cosmosaccount"
	"gopkg.in/yaml.v3"
)

func UpdateNestedYAML(node *yaml.Node, path []string, value interface{}) error {
	if node.Kind == yaml.DocumentNode {
		if len(node.Content) == 0 {
			return fmt.Errorf("empty document node")
		}
		return UpdateNestedYAML(node.Content[0], path, value)
	}

	if len(path) == 0 {
		return setNodeValue(node, value)
	}

	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("expected a mapping node, got %v", node.Kind)
	}

	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == path[0] {
			return UpdateNestedYAML(node.Content[i+1], path[1:], value)
		}
	}

	fmt.Println(path)
	// If the path doesn't exist, create it
	// Create a new key node
	newKeyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: path[0],
		Tag:   "!!str",
	}
	node.Content = append(node.Content, newKeyNode)

	// Determine the kind of the new value node
	var newValueNode *yaml.Node
	if len(path) == 1 {
		// If this is the last element in the path, set the value
		newValueNode = &yaml.Node{
			Kind: yaml.ScalarNode,
			Tag:  "!!str", // You can adjust the tag based on the type of `value`
		}
	} else {
		// Otherwise, create a new mapping node for the next level
		newValueNode = &yaml.Node{
			Kind: yaml.MappingNode,
		}
	}

	node.Content = append(node.Content, newValueNode)
	return UpdateNestedYAML(newValueNode, path[1:], value)
}

func setNodeValue(node *yaml.Node, value interface{}) error {
	switch v := value.(type) {
	case string:
		node.Kind = yaml.ScalarNode
		node.Tag = "!!str"
		node.Value = v
	case int:
		node.Kind = yaml.ScalarNode
		node.Tag = "!!int"
		node.Value = fmt.Sprintf("%d", v)
	case float32:
		node.Kind = yaml.ScalarNode
		node.Tag = "!!float"
		node.Value = fmt.Sprintf("%f", v)
	case float64:
		node.Kind = yaml.ScalarNode
		node.Tag = "!!float"
		node.Value = fmt.Sprintf("%f", v)
	case bool:
		node.Kind = yaml.ScalarNode
		node.Tag = "!!bool"
		node.Value = fmt.Sprintf("%t", v)
	default:
		return fmt.Errorf("unsupported value type: %T", value)
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

	LogLevel    string      `mapstructure:"log_level"`
	SlackConfig slackConfig `mapstructure:"slack"`
	SkipRefund  bool        `mapstructure:"skip_refund"`
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
