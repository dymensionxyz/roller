package relayer

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	toml "github.com/BurntSushi/toml"
	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/relayer"
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/rollapp"
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

	// Create root context with signal handling
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Create separate contexts for each command
	hubCtx, hubCancel := context.WithCancel(ctx)
	raCtx, raCancel := context.WithCancel(ctx)
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
					"[Hub] Starting flush for range %d -> %d\n",
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
					pterm.Error.Printf("[Hub] Flush command failed: %v\n", err)
					hubCancel()
					return
				}

				pterm.Info.Printf(
					"[Hub] Flush completed for range %d -> %d\n",
					startHeight,
					endHeight,
				)

				// Update the last hub flush height
				currentCfg.LastHubFlushHeight += currentCfg.FlushRange
				if err := writeFlushConfig(rrhf, currentCfg); err != nil {
					pterm.Error.Printf("[Hub] Failed to update flush height: %v\n", err)
					return
				}

				// Log completion of current range
				pterm.Info.Printf("[Hub] Moving to next range...\n")
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

				// Get current rollapp height
				blockInfo, err := rollapp.GetCurrentHeight()
				if err != nil {
					pterm.Error.Printf("failed to get current rollapp height: %v\n", err)
					return
				}

				currentHeight, err := strconv.Atoi(blockInfo.Block.Header.Height)
				if err != nil {
					pterm.Error.Printf("failed to parse current height: %v\n", err)
					return
				}

				startHeight := currentCfg.LastRaFlushHeight
				endHeight := startHeight + currentCfg.FlushRange

				if endHeight > currentHeight {
					pterm.Info.Printf(
						"[RollApp] Target end height %d is greater than current height %d, setting LastRaFlushHeight to current height\n",
						endHeight,
						currentHeight,
					)
					// Update config to current height and exit
					currentCfg.LastRaFlushHeight = currentHeight
					endHeight = currentHeight
					if err := writeFlushConfig(rrhf, currentCfg); err != nil {
						pterm.Error.Printf("[RollApp] Failed to update flush height: %v\n", err)
						return
					}
				}

				pterm.Info.Printf(
					"[RollApp] Starting flush for range %d -> %d\n",
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
					pterm.Error.Printf("[RollApp] Flush command failed: %v\n", err)
					raCancel()
					return
				}

				pterm.Info.Printf(
					"[RollApp] Flush completed for range %d -> %d\n",
					startHeight,
					endHeight,
				)

				// Update the last rollapp flush height
				currentCfg.LastRaFlushHeight = endHeight
				if err := writeFlushConfig(rrhf, currentCfg); err != nil {
					pterm.Error.Printf("[RollApp] Failed to update flush height: %v\n", err)
					return
				}

				// If we've caught up to current height, sleep for a bit and check again
				if endHeight >= currentHeight {
					pterm.Info.Printf(
						"[RollApp] Caught up to current height %d, waiting for new blocks...\n",
						currentHeight,
					)
					select {
					case <-raCtx.Done():
						return
					case <-time.After(10 * time.Second):
						continue
					}
				}

				// Log progress and continue to next range
				pterm.Info.Printf("[RollApp] Moving to next range...\n")
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
	// Try to load existing config first
	var config RollerRelayerHelperConfig

	_, err := os.Stat(rrhf)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			pterm.Error.Println("failed to check relayer helper file")
			return nil, err
		}

		// File doesn't exist, create new config
		pterm.Info.Printfln("%s does not exist, creating new config", rrhf)
		err := os.MkdirAll(filepath.Dir(rrhf), 0o755)
		if err != nil {
			pterm.Error.Printfln("failed to create directory for %s: %v", rrhf, err)
			return nil, err
		}

		// Create the file
		f, err := os.Create(rrhf)
		if err != nil {
			pterm.Error.Printf("failed to create config file: %v\n", err)
			return nil, err
		}
		f.Close()

		hubFlushHeight, err := sequencer.GetFirstStateUpdateHeight(raID, hd.RpcUrl, hd.ID)
		if err != nil {
			pterm.Error.Println("failed to retrieve the height of the first state update:", err)
			return nil, err
		}

		// Initialize new config
		config = RollerRelayerHelperConfig{
			LastRaFlushHeight:  1,
			LastHubFlushHeight: hubFlushHeight,
			FlushRange:         10_000,
		}

		// Write initial config
		if err := writeFlushConfig(rrhf, &config); err != nil {
			pterm.Error.Printf("failed to write initial config: %v\n", err)
			return nil, err
		}

		return &config, nil
	}

	// Load existing config
	cfg, err := tomlconfig.Load(rrhf)
	if err != nil {
		pterm.Error.Printf("failed to load relayer helper config: %v\n", err)
		return nil, err
	}

	err = toml.Unmarshal(cfg, &config)
	if err != nil {
		pterm.Error.Printf("failed to unmarshal relayer helper config: %v\n", err)
		return nil, err
	}

	return &config, nil
}
