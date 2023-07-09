package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

const (
	networksFileName = "networks.yaml"

	StagingHubName = "devnet"
	LocalHubName   = "local"
)

type Config struct {
	Hubs map[string]HubData `yaml:"Hubs"`
}

type HubData struct {
	API_URL string `yaml:"API_URL"`
	ID      string `yaml:"ID"`
	RPC_URL string `yaml:"RPC_URL"`
}

var Hubs = map[string]HubData{
	StagingHubName: {
		API_URL: "https://dymension.devnet.api.silknodes.io:443",
		ID:      "devnet_304-1",
		RPC_URL: "https://dymension.devnet.rpc.silknodes.io:443",
	},
	LocalHubName: {
		API_URL: "http://localhost:1318",
		ID:      "dymension_100-1",
		RPC_URL: "http://localhost:36657",
	},
}

func LoadNetworksFromFile() {
	// Read the YAML file
	file, err := ioutil.ReadFile(networksFileName)
	if err != nil {
		fmt.Println("failed to read networks from $s. using defaults", networksFileName)
		return
	}

	// Unmarshal the YAML into the config struct
	var config Config
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		fmt.Println("failed to unmarshal networks from $s. using defautls", networksFileName)
		return
	}

	Hubs = config.Hubs
}
