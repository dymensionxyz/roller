package sequencer

import (
	"github.com/pelletier/go-toml"
	"os"
)

func EnableDymintMetrics(root string) error {
	dymintTomlPath := GetDymintFilePath(root)
	config, err := toml.LoadFile(dymintTomlPath)
	if err != nil {
		return err
	}
	config.Set("instrumentation.prometheus", true)
	config.Set("instrumentation.prometheus_listen_addr", ":2112")
	file, err := os.Create(dymintTomlPath)
	if err != nil {
		return err
	}
	_, err = file.WriteString(config.String())
	if err != nil {
		return err
	}
	return nil
}
