package relayer

import cosmossdkmath "cosmossdk.io/math"

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
	Paths struct {
		HubRollapp struct {
			Dst struct {
				ChainID string `yaml:"chain-id"`
			} `yaml:"dst"`
			Src struct {
				ChainID string `yaml:"chain-id"`
			} `yaml:"src"`
		} `yaml:"hub-rollapp"`
	} `yaml:"paths"`
}
