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
	"strings"
	"sync"
	"syscall"

	toml "github.com/BurntSushi/toml"
	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/relayer"
	"github.com/dymensionxyz/roller/sequencer"
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

	// flushRange handles a single range of blocks for either hub or rollapp
	flushRange := func(ctx context.Context, startHeight, endHeight int, isHub bool) error {
		chainID := hd.ID
		prefix := "[Hub]"
		if !isHub {
			chainID = raID
			prefix = "[RollApp]"
		}

		pterm.Info.Printf(
			"%s Starting flush for range %d -> %d\n",
			prefix,
			startHeight,
			endHeight,
		)

		flushCmd := getFlushCmd(
			rlyConfigDir,
			chainID,
			startHeight,
			endHeight-startHeight,
		)
		fmt.Println(flushCmd.String())

		doneChan := make(chan error, 1)
		outChan := make(chan []byte, 100)

		// Start command execution in a goroutine
		go func() {
			cmd := exec.CommandContext(ctx, flushCmd.Path, flushCmd.Args...)
			cmd.Env = flushCmd.Env

			// Set up pipes for stdout and stderr
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				doneChan <- fmt.Errorf("failed to create stdout pipe: %v", err)
				return
			}
			stderr, err := cmd.StderrPipe()
			if err != nil {
				doneChan <- fmt.Errorf("failed to create stderr pipe: %v", err)
				return
			}

			// Start the command
			if err := cmd.Start(); err != nil {
				doneChan <- fmt.Errorf("failed to start command: %v", err)
				return
			}

			// Read stdout and stderr in separate goroutines
			go func() {
				buf := make([]byte, 1024)
				for {
					n, err := stdout.Read(buf)
					if n > 0 {
						outChan <- buf[:n]
					}
					if err != nil {
						break
					}
				}
			}()

			go func() {
				buf := make([]byte, 1024)
				for {
					n, err := stderr.Read(buf)
					if n > 0 {
						outChan <- buf[:n]
					}
					if err != nil {
						break
					}
				}
			}()

			// Wait for command completion
			if err := cmd.Wait(); err != nil {
				doneChan <- err
				return
			}
			close(outChan)
			doneChan <- nil
		}()

		// Process output and watch for skip signal
		var output string
		for {
			select {
			case data, ok := <-outChan:
				if !ok {
					// Command completed normally
					pterm.Info.Printf("%s Moving to next range...\n", prefix)
					return nil
				}
				output += string(data)
				if strings.Contains(output, "Parsed stuck packet height, skipping to current") {
					pterm.Info.Printf("%s Range complete, skipping to next range\n", prefix)
					return nil
				}
			case err := <-doneChan:
				if err != nil {
					pterm.Error.Printf("%s Flush command failed: %v\n", prefix, err)
					return err
				}
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

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

				startHeight := currentCfg.LastHubFlushHeight
				endHeight := startHeight + currentCfg.FlushRange

				// Flush this range
				err = flushRange(hubCtx, startHeight, endHeight, true)
				if err != nil {
					return
				}

				// Update the last hub flush height
				currentCfg.LastHubFlushHeight = endHeight
				if err := writeFlushConfig(rrhf, currentCfg); err != nil {
					pterm.Error.Printf("[Hub] Failed to update flush height: %v\n", err)
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

				// If we've caught up to current height, exit
				if startHeight >= currentHeight {
					pterm.Info.Printf(
						"[RollApp] Caught up to current height %d, exiting\n",
						currentHeight,
					)
					return
				}

				// Adjust end height if it would exceed current height
				if endHeight > currentHeight {
					pterm.Info.Printf(
						"[RollApp] Adjusting end height from %d to current height %d\n",
						endHeight,
						currentHeight,
					)
					endHeight = currentHeight
				}

				// Flush this range
				err = flushRange(raCtx, startHeight, endHeight, false)
				if err != nil {
					return
				}

				// Update the last rollapp flush height
				currentCfg.LastRaFlushHeight = endHeight
				if err := writeFlushConfig(rrhf, currentCfg); err != nil {
					pterm.Error.Printf("[RollApp] Failed to update flush height: %v\n", err)
					return
				}

				// If we've caught up to current height, exit
				if endHeight >= currentHeight {
					pterm.Info.Printf(
						"[RollApp] Caught up to current height %d, exiting\n",
						currentHeight,
					)
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
			FlushRange:         100,
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
