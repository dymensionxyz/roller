package relayer

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	toml "github.com/BurntSushi/toml"
	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/relayer"
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
)

type RollerRelayerHelperConfig struct {
	FlushRange         int `toml:"flush_range"`
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
	hd := rlyConfig.HubDataFromRelayerConfig()
	raID := rlyConfig.Paths.HubRollapp.Dst.ChainID

	flushCfg, err := getFlushConfig(rrhf, raID, *hd)
	if err != nil {
		pterm.Error.Println("failed to handle flusher config:", err)
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
	pterm.Info.Printfln(
		"flush interval: %v",
		flushCfg.FlushRange,
	)

	pterm.Info.Println("chains to flush:")
	for _, v := range chainsToFlush {
		fmt.Println(v)
	}

	// Create separate contexts for each command
	hubCtx, hubCancel := context.WithCancel(context.Background())
	raCtx, raCancel := context.WithCancel(context.Background())
	defer hubCancel()
	defer raCancel()

	var wg sync.WaitGroup
	wg.Add(2)

	// Start hub flush goroutine
	go func() {
		defer wg.Done()
		for {
			select {
			case <-hubCtx.Done():
				return
			default:
				currentCfg, err := getFlushConfig(rrhf, raID, *hd)
				if err != nil {
					pterm.Error.Printf("failed to get current flush config for hub: %v\n", err)
					return
				}

				// Skip if height is 0 on first run
				if currentCfg.LastHubFlushHeight == 0 {
					currentCfg.LastHubFlushHeight = 1
					if err := writeFlushConfig(rrhf, currentCfg); err != nil {
						pterm.Error.Printf("failed to initialize hub flush height: %v\n", err)
						return
					}
				}

				startHeight := currentCfg.LastHubFlushHeight
				endHeight := startHeight + currentCfg.FlushRange

				pterm.Info.Printf(
					"Starting hub flush from height %d to %d\n",
					startHeight,
					endHeight,
				)

				hubFlushCmd := getFlushCmd(
					rlyConfigDir,
					hd.ID,
					startHeight,
					currentCfg.FlushRange,
				)

				doneChan := make(chan error, 1)
				err = bash.ExecCmdFollow(doneChan, hubCtx, hubFlushCmd, nil)
				if err != nil {
					pterm.Error.Printf("hub flush command failed: %v\n", err)
					hubCancel()
					return
				}

				pterm.Info.Printf(
					"Hub flush completed for range %d to %d\n",
					startHeight,
					endHeight,
				)

				// Update the last hub flush height
				currentCfg.LastHubFlushHeight += currentCfg.FlushRange
				if err := writeFlushConfig(rrhf, currentCfg); err != nil {
					pterm.Error.Printf("failed to update hub flush height: %v\n", err)
					return
				}
			}
		}
	}()

	// Start rollapp flush goroutine
	go func() {
		defer wg.Done()
		for {
			select {
			case <-raCtx.Done():
				return
			default:
				currentCfg, err := getFlushConfig(rrhf, raID, *hd)
				if err != nil {
					pterm.Error.Printf("failed to get current flush config for rollapp: %v\n", err)
					return
				}

				// Skip if height is 0 on first run
				if currentCfg.LastRaFlushHeight == 0 {
					currentCfg.LastRaFlushHeight = 1
					if err := writeFlushConfig(rrhf, currentCfg); err != nil {
						pterm.Error.Printf("failed to initialize rollapp flush height: %v\n", err)
						return
					}
				}

				startHeight := currentCfg.LastRaFlushHeight
				endHeight := startHeight + currentCfg.FlushRange

				pterm.Info.Printf(
					"Starting rollapp flush from height %d to %d\n",
					startHeight,
					endHeight,
				)

				raFlushCmd := getFlushCmd(
					rlyConfigDir,
					raID,
					startHeight,
					currentCfg.FlushRange,
				)

				doneChan := make(chan error, 1)
				err = bash.ExecCmdFollow(doneChan, raCtx, raFlushCmd, nil)
				if err != nil {
					pterm.Error.Printf("rollapp flush command failed: %v\n", err)
					raCancel()
					return
				}

				pterm.Info.Printf(
					"Rollapp flush completed for range %d to %d\n",
					startHeight,
					endHeight,
				)

				// Update the last rollapp flush height
				currentCfg.LastRaFlushHeight += currentCfg.FlushRange
				if err := writeFlushConfig(rrhf, currentCfg); err != nil {
					pterm.Error.Printf("failed to update rollapp flush height: %v\n", err)
					return
				}
			}
		}
	}()

	wg.Wait()
}

// nolint unused
func writeFlushConfig(configPath string, config *RollerRelayerHelperConfig) error {
	f, err := os.OpenFile(configPath, os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	return toml.NewEncoder(f).Encode(config)
}

func getFlushCmd(rlyConfigDir, chain string, startHeight, r int) *exec.Cmd {
	endHeight := startHeight + r

	shStr := fmt.Sprintf("%d", startHeight)
	ehStr := fmt.Sprintf("%d", endHeight)

	cmd := exec.Command(
		consts.Executables.Relayer,
		"tx",
		"flush",
		"--stuck-packet-chain-id",
		chain,
		"--stuck-packet-height-start",
		shStr,
		"--stuck-packet-height-end",
		ehStr,
		"hub-rollapp",
		"--home",
		rlyConfigDir,
	)

	return cmd
}

func getFlushConfig(rrhf, raID string, hd consts.HubData) (*RollerRelayerHelperConfig, error) {
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

			hubFlushHeight, err := sequencer.GetFirstStateUpdateHeight(raID, hd.RpcUrl, hd.ID)
			if err != nil {
				pterm.Error.Println("failed to retrieve the height of the first state update:", err)
				return nil, err
			}

			// Load existing config
			var config RollerRelayerHelperConfig
			if _, err := toml.DecodeFile(rrhf, &config); err != nil {
				return nil, err
			}

			// Update values
			config.LastRaFlushHeight = 1
			config.LastHubFlushHeight = hubFlushHeight
			config.FlushRange = 10_000

			// Write back to file
			f, err := os.OpenFile(rrhf, os.O_WRONLY|os.O_TRUNC, 0o644)
			if err != nil {
				return nil, err
			}
			defer f.Close()

			return &config, toml.NewEncoder(f).Encode(config)
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
