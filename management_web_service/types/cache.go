package types

import (
	"fmt"
	"os"
	"sync"
	"time"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

const eIbcClientLogSize = 1000

// Cache is a thread-safe cache for the web service
type Cache struct {
	mu                      sync.RWMutex
	eIbcClientLog           []string
	eIbcClientProcess       *os.Process
	lastStartEIbcClient     time.Time
	eIbcClientDenom         string
	eIbcClientMinFeePercent float64
	denomsMetadata          map[string]banktypes.Metadata
}

func (c *Cache) CanStartEIbcClientProcessID() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.eIbcClientProcess != nil {
		return fmt.Errorf("eIBC client is already running at PID: %d", c.eIbcClientProcess.Pid)
	}

	if time.Since(c.lastStartEIbcClient) < 15*time.Second {
		return fmt.Errorf("eIBC client was started too recently")
	}

	return nil
}

// GetEIbcClientProcessID returns the eIBC-client process ID
func (c *Cache) GetEIbcClientProcessID() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.eIbcClientProcess == nil {
		return 0
	}

	return c.eIbcClientProcess.Pid
}

// SetEIbcClientProcess sets the eIBC-client process
func (c *Cache) SetEIbcClientProcess(p *os.Process) {
	if p != nil {
		c.AppendEIbcClientLog(fmt.Sprintf("eIBC client started, pid: %d", p.Pid))
		c.lastStartEIbcClient = time.Now().UTC()
	} else {
		c.AppendEIbcClientLog("eIBC client stopped")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.eIbcClientProcess = p
}

// GetEIbcClientProcess returns the eIBC-client process
func (c *Cache) GetEIbcClientProcess() *os.Process {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.eIbcClientProcess
}

// AppendEIbcClientLog appends a new eIBC-client log
func (c *Cache) AppendEIbcClientLog(log string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.eIbcClientLog = append(c.eIbcClientLog, fmt.Sprintf("%s | %s", time.Now().UTC().Format("Jan 02 15:04"), log))
	if len(c.eIbcClientLog) > eIbcClientLogSize+100 /*avoid malloc everytime*/ {
		c.eIbcClientLog = c.eIbcClientLog[len(c.eIbcClientLog)-eIbcClientLogSize:]
	}
}

// GetEIbcClientLog returns the last eIBC-client logs
func (c *Cache) GetEIbcClientLog() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	currentSize := len(c.eIbcClientLog)
	size := currentSize
	if size > eIbcClientLogSize {
		size = eIbcClientLogSize
	}
	res := make([]string, 0, size)
	for i := currentSize - 1; i >= currentSize-size; i-- {
		res = append(res, c.eIbcClientLog[i])
	}
	return res
}

func (c *Cache) SetEIbcClientArgs(denom string, minFeePercent float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.eIbcClientDenom = denom
	c.eIbcClientMinFeePercent = minFeePercent
}

func (c *Cache) GetEIbcClientArgs() (denom string, minFeePercent float64) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	denom = c.eIbcClientDenom
	minFeePercent = c.eIbcClientMinFeePercent
	return
}

func (c *Cache) SetDenomsMetadata(denomsMetadata map[string]banktypes.Metadata) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.denomsMetadata = denomsMetadata
}

func (c *Cache) GetDenomsMetadata() map[string]banktypes.Metadata {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.denomsMetadata == nil {
		return map[string]banktypes.Metadata{}
	}

	return c.denomsMetadata
}
