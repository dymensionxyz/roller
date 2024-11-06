package keys

import (
	"runtime"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
)

func KeyringBackendFromEnv(env string) consts.SupportedKeyringBackend {
	switch env {
	case "mock", "playground":
		return consts.SupportedKeyringBackends.Test
	case "custom":
		krBackends := []string{"test"}
		if runtime.GOOS != "darwin" {
			krBackends = append(krBackends, "os")
		}
		keyringBackend, _ := pterm.DefaultInteractiveSelect.WithOptions(krBackends).Show()
		return consts.SupportedKeyringBackend(keyringBackend)
	default:
		if runtime.GOOS != "darwin" {
			return consts.SupportedKeyringBackends.OS
		}
		return consts.SupportedKeyringBackends.Test
	}
}
