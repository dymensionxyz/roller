package relayer

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v3"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/roller"
)

type Relayer struct {
	RollerHome     string
	RelayerHome    string
	ConfigFilePath string

	Rollapp consts.RollappData
	Hub     consts.HubData
	// channels
	SrcChannel string
	DstChannel string
	// connections
	SrcConnectionID string
	DstConnectionID string
	// clients
	SrcClientID string
	DstClientID string

	Config *Config

	logger *log.Logger
}

func NewRelayer(home string, raData consts.RollappData, hd consts.HubData) *Relayer {
	relayerHome := GetHomeDir(home)
	relayerConfigPath := GetConfigFilePath(relayerHome)
	return &Relayer{
		RollerHome:     home,
		RelayerHome:    relayerHome,
		ConfigFilePath: relayerConfigPath,

		Rollapp: raData,
		Hub:     hd,

		logger: log.New(io.Discard, "", 0),
	}
}

func (r *Relayer) SetLogger(logger *log.Logger) {
	r.logger = logger
}

// TODO: review the servicemanager.Service implementation
func (r *Relayer) GetRelayerStatus(roller.RollappConfig) string {
	if r.ChannelReady() {
		return fmt.Sprintf(
			"Active Channels:\nrollapp: %s\n<->\nhub: %s",
			r.SrcChannel,
			r.DstChannel,
		)
	}
	bytes, err := os.ReadFile(r.StatusFilePath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "Starting..."
		}
	}
	return string(bytes)
}

func (r *Relayer) WriteRelayerStatus(status string) error {
	// nolint:gofumpt
	return os.WriteFile(r.StatusFilePath(), []byte(status), 0o644)
}

func (r *Relayer) StatusFilePath() string {
	return filepath.Join(r.RollerHome, consts.ConfigDirName.Relayer, "relayer_status.txt")
}

type ConnectionChannels struct {
	Src string
	Dst string
}

func (c *Config) Load(rlyConfigPath string) error {
	data, err := os.ReadFile(rlyConfigPath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, c)
	if err != nil {
		return err
	}

	return nil
}
