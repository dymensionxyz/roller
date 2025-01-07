package config

import (
	"strings"

	"github.com/pterm/pterm"
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
