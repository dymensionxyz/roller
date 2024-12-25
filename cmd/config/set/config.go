package set

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	toml "github.com/pelletier/go-toml"
	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/rollapp/setup"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/roller"
	"github.com/dymensionxyz/roller/utils/sequencer"
	sequencerutils "github.com/dymensionxyz/roller/utils/sequencer"
	servicemanager "github.com/dymensionxyz/roller/utils/service_manager"
)

func verifyMinimumGasPrice(
	rollerData roller.RollappConfig,
	value string,
) (string, *sequencer.DenomMetadata, error) {
	if value == "" {
		return "", nil, fmt.Errorf("minimum gas price should be provided")
	}

	if _, err := strconv.ParseFloat(value, 64); err == nil {
		return "", nil, fmt.Errorf(
			"minimum gas price should be in the format <number><denom>, e.g. 100000amock or 100000ibc/FECACB927EB3102CCCB240FFB3B6FCCEEB8D944C6FEA8DFF079650FEFF59781D ( for dym )",
		)
	}

	// Find the first non-digit character to separate number from denomination
	i := 0
	for i < len(value) && unicode.IsDigit(rune(value[i])) {
		i++
	}
	if i == 0 || i == len(value) {
		return "", nil, fmt.Errorf(
			"invalid minimum gas price format: must start with a number and include denomination",
		)
	}

	denom := value[i:]
	if !strings.HasPrefix(denom, "a") &&
		denom != "ibc/FECACB927EB3102CCCB240FFB3B6FCCEEB8D944C6FEA8DFF079650FEFF59781D" {
		return "", nil, fmt.Errorf(
			"invalid denom: minimum gas price denomination must start with 'a' (e.g., amock) or ibc/FECACB927EB3102CCCB240FFB3B6FCCEEB8D944C6FEA8DFF079650FEFF59781D",
		)
	}

	sgd, err := setup.SupportedGasDenoms(rollerData)
	if err != nil {
		return "", nil, err
	}

	if _, ok := sgd[denom]; !ok {
		return "", nil, fmt.Errorf("invalid denom: minimum gas price denomination not supported")
	}

	var display string
	var dm sequencer.DenomMetadata
	if denom == "ibc/FECACB927EB3102CCCB240FFB3B6FCCEEB8D944C6FEA8DFF079650FEFF59781D" {
		display = "dym"
		dm = sequencer.DenomMetadata{
			Display:  display,
			Base:     denom,
			Exponent: 18,
		}
		return value, &dm, nil
	} else {
		return value, nil, nil
	}
}

func setMinimumGasPrice(rollerData roller.RollappConfig, value string) error {
	amount, dm, err := verifyMinimumGasPrice(rollerData, value)
	if err != nil {
		return err
	}

	hubSeqKC := keys.KeyConfig{
		Dir:            consts.ConfigDirName.HubKeys,
		ID:             consts.KeysIds.HubSequencer,
		ChainBinary:    consts.Executables.Dymension,
		Type:           consts.SDK_ROLLAPP,
		KeyringBackend: rollerData.KeyringBackend,
	}

	seqAddrInfo, err := hubSeqKC.Info(rollerData.Home)
	if err != nil {
		return err
	}
	seqAddrInfo.Address = strings.TrimSpace(seqAddrInfo.Address)

	metadata, err := sequencer.GetMetadata(seqAddrInfo.Address, rollerData.HubData)
	if err != nil {
		return err
	}

	metadata.FeeDenom = dm
	metadata.GasPrice = amount

	appConfigFilePath := filepath.Join(
		sequencerutils.GetSequencerConfigDir(rollerData.Home),
		"app.toml",
	)
	appCfg, err := toml.LoadFile(appConfigFilePath)
	if err != nil {
		return fmt.Errorf("failed to load %s: %v", appConfigFilePath, err)
	}
	appCfg.Set("minimum-gas-prices", value)
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
