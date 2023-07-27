package sequencer

import "github.com/pelletier/go-toml"

func EnableDymintMetrics(root string) error {
	dymintTomlPath := GetDymintFilePath(root)
	config, err := toml.LoadFile(dymintTomlPath)
	if err != nil {
		return err
	}
	config.Set("instrumentation.prometheus", true)
	return nil
}
