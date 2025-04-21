package set

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/roller"
	sequencerutils "github.com/dymensionxyz/roller/utils/sequencer"
	servicemanager "github.com/dymensionxyz/roller/utils/service_manager"
)

func setMinimumGasPrice(cfg roller.RollappConfig, value string) error {
	appConfigFilePath := filepath.Join(sequencerutils.GetSequencerConfigDir(cfg.Home), "app.toml")
	err := tomlconfig.UpdateFieldInFile(appConfigFilePath, "minimum-gas-prices", value)
	if err != nil {
		return err
	}
	return nil
}

func setBlockTime(cfg roller.RollappConfig, value string) error {
	dymintTomlPath := sequencerutils.GetDymintFilePath(cfg.Home)

	pterm.Info.Println("updating block time")
	_, err := time.ParseDuration(value)
	if err != nil {
		return fmt.Errorf(
			"invalid duration format, expected format like '1h0m0s' or '2m2s': %v",
			err,
		)
	}

	if duration, err := time.ParseDuration(value); err != nil || duration < 5*time.Second {
		return fmt.Errorf("optimal block time should be bigger then 5 seconds")
	}

	updates := map[string]any{
		"max_idle_time":     value,
		"batch_submit_time": value,
	}

	for k, v := range updates {
		err := tomlconfig.UpdateFieldInFile(dymintTomlPath, k, v)
		if err != nil {
			return err
		}
	}

	pterm.Info.Println("block time updated, restarting rollapp")

	err = servicemanager.RestartSystemServices([]string{"rollapp"}, cfg.Home)
	if err != nil {
		return err
	}

	return nil
}
