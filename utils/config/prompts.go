package config

import (
	"strings"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/roller"
)

func PromptVmType() string {
	vmtypes := []string{"evm", "wasm"}
	vmtype, _ := pterm.DefaultInteractiveSelect.
		WithDefaultText("select the rollapp VM type you want to initialize for").
		WithOptions(vmtypes).
		Show()

	return vmtype
}

func PromptRaID() string {
	raID, _ := pterm.DefaultInteractiveTextInput.WithDefaultText("Please enter the RollApp ID").
		Show()

	return strings.TrimSpace(raID)
}

func PromptEnvironment() string {
	envs := []string{"playground", "blumbus", "custom", "mainnet"}
	env, _ := pterm.DefaultInteractiveSelect.
		WithDefaultText(
			"select the environment you want to initialize relayer for",
		).
		WithOptions(envs).
		Show()

	return strings.TrimSpace(env)
}

func PromptCustomHubEndpoint(rollerConfig roller.RollappConfig) roller.RollappConfig {
	if rollerConfig.HubData.Environment != "mainnet" {
		return rollerConfig
	}

	var rpcEndpoint string

	if rollerConfig.HubData.RpcUrl == consts.MainnetHubData.RpcUrl || rollerConfig.HubData.RpcUrl == "" {
		for {
			rpcEndpoint, _ = pterm.DefaultInteractiveTextInput.WithDefaultText("We recommend using a private RPC endpoint for the hub. Please provide the hub rpc endpoint to use. You can obtain one here: https://blastapi.io/chains/dymension").
				Show()

			isValidUrl := IsValidURL(rpcEndpoint)
			if isValidUrl {
				break
			}
		}

		if rpcEndpoint != "" {
			rollerConfig.HubData.RpcUrl = rpcEndpoint
			rollerConfig.HubData.ArchiveRpcUrl = rpcEndpoint
		}
	}

	return rollerConfig
}
