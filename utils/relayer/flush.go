package relayer

import (
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/naoina/toml"
	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/relayer"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
)

type RollerRelayerHelperConfig struct {
	LastHubFlushHeight int `toml:"last_hub_flush_height"`
	LastRaFlushHeight  int `toml:"last_ra_flush_height"`
}

func Flush(home string) {
	rlyConfigDir := filepath.Join(
		home,
		consts.ConfigDirName.Relayer,
	)
	rlyCfgPath := filepath.Join(
		rlyConfigDir,
		"config",
		"config.yaml",
	)

	rrhf := filepath.Join(
		rlyConfigDir,
		"roller-relayer-helper.toml",
	)

	var rlyConfig relayer.Config
	chainsToFlush, err := rlyConfig.GetChains(rlyCfgPath)
	if err != nil {
		pterm.Error.Println("failed to retrieve chains to run flush for: ", err)
		return
	}

	flushCfg, err := getFlushConfig(rrhf)
	if err != nil {
		pterm.Error.Println("failed to handle flusher config")
		return
	}

	pterm.Info.Printfln(
		"retrieved RollApp height to start flushing from: %v",
		flushCfg.LastRaFlushHeight,
	)
	pterm.Info.Printfln(
		"retrieved Hub height to start flushing from: %v",
		flushCfg.LastHubFlushHeight,
	)

	pterm.Info.Println("chains to flush:", chainsToFlush)
}

// nolint unused
func getFlushCmd(rlyConfigDir, startHeight, endHeight, chain string) *exec.Cmd {
	cmd := exec.Command(
		consts.Executables.Relayer,
		"tx",
		"flush",
		"--stuck-packet-chain-id",
		chain,
		"--stuck-packet-height-start",
		startHeight,
		"--stuck-packet-height-end",
		endHeight,
		"hub-rollapp",
		"--home",
		rlyConfigDir,
	)

	return cmd
}

func getFlushConfig(rrhf string) (*RollerRelayerHelperConfig, error) {
	_, err := os.Stat(rrhf)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			pterm.Info.Printfln("%s does not exist", rrhf)
			err := os.MkdirAll(filepath.Dir(rrhf), 0o755)
			if err != nil {
				pterm.Error.Printfln("failed to create directory for %s", rrhf)
			}

			_, err = os.Create(rrhf)
			if err != nil {
				pterm.Error.Printfln("failed to create %s", rrhf)
			}

			updates := map[string]any{
				"last_ra_flush_height":  0,
				"last_hub_flush_height": 0,
			}

			for k, v := range updates {
				err = tomlconfig.UpdateFieldInFile(
					rrhf,
					k,
					v,
				)
				if err != nil {
					pterm.Error.Println("failed to update relayer helper config: ", err)
				}
			}
		} else {
			pterm.Error.Println("failed to check relayer helper file")
		}
	}

	var rrhc RollerRelayerHelperConfig
	cfg, err := tomlconfig.Load(rrhf)
	if err != nil {
		pterm.Error.Println("failed to load relayer helper config")
	}

	err = toml.Unmarshal(cfg, &rrhc)
	if err != nil {
		pterm.Error.Println("failed to unmarshal relayer helper config")
	}

	return &rrhc, nil
}
