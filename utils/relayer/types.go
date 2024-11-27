package relayer

import (
	cosmossdkmath "cosmossdk.io/math"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
)

var oneDayRelayPrice, _ = cosmossdkmath.NewIntFromString(
	"2000000000000000000",
) // 2000000000000000000 = 2dym

type Channels struct {
	Channels []struct {
		State        string `json:"state"`
		Ordering     string `json:"ordering"`
		Counterparty struct {
			PortId    string `json:"port_id"`
			ChannelId string `json:"channel_id"`
		} `json:"counterparty"`
		ConnectionHops []string `json:"connection_hops"`
		Version        string   `json:"version"`
		PortId         string   `json:"port_id"`
		ChannelId      string   `json:"channel_id"`
	} `json:"channels"`
	Pagination struct {
		NextKey interface{} `json:"next_key"`
		Total   string      `json:"total"`
	} `json:"pagination"`
	Height struct {
		RevisionNumber string `json:"revision_number"`
		RevisionHeight string `json:"revision_height"`
	} `json:"height"`
}

// Config struct represents the paths section inside the relayer
// configuration file
type Config struct {
	Chains map[string]initconfig.RelayerFileChainConfig `yaml:"chains"`
	Paths  *struct {
		HubRollapp *struct {
			Dst *struct {
				ChainID      string `yaml:"chain-id"`
				ClientID     string `yaml:"client-id"`
				ConnectionID string `yaml:"connection-id"`
			} `yaml:"dst"`
			Src *struct {
				ChainID      string `yaml:"chain-id"`
				ClientID     string `yaml:"client-id"`
				ConnectionID string `yaml:"connection-id"`
			} `yaml:"src"`
			SrcChannelFilter *struct {
				ChannelList []string `yaml:"channel-list"`
				Rule        string   `yaml:"rule"`
			} `yaml:"src-channel-filter"`
		} `yaml:"hub-rollapp"`
	} `yaml:"paths"`
}
