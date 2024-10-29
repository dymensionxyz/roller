package keys

import "github.com/pterm/pterm"

func KeyringBackendFromEnv(env string) string {
	var keyringBackend string
	if env == "mock" || env == "playground" {
		keyringBackend = "test"
	} else if env == "custom" {
		krBackends := []string{"os", "test"}
		keyringBackend, _ = pterm.DefaultInteractiveSelect.WithOptions(krBackends).Show()
	} else {
		keyringBackend = "os"
	}

	return keyringBackend
}
