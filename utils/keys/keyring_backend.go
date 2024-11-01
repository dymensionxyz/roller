package keys

import (
	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
)

func KeyringBackendFromEnv(env string) consts.SupportedKeyringBackend {
	switch env {
	case "mock", "playground":
		return consts.SupportedKeyringBackends.Test
	case "custom":
		krBackends := []string{"os", "test"}
		keyringBackend, _ := pterm.DefaultInteractiveSelect.WithOptions(krBackends).Show()
		return consts.SupportedKeyringBackend(keyringBackend)
	default:
		return consts.SupportedKeyringBackends.OS
	}
}
