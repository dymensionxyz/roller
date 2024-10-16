package config

import "github.com/pterm/pterm"

func PromptVmType() string {
	vmtypes := []string{"evm", "wasm"}
	vmtype, _ := pterm.DefaultInteractiveSelect.
		WithDefaultText("select the rollapp VM type you want to initialize for").
		WithOptions(vmtypes).
		Show()

	return vmtype
}
