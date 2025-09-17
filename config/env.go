package config

import (
	"github.com/caarlos0/env/v11"
)

type EnvConfig struct {
	RollappCommit       string `env:"ROLLER_RA_COMMIT"`
	RollappGenesis      string `env:"ROLLER_RA_GENESIS"`
	SkipCelestiaBinary  bool   `env:"ROLLER_SKIP_CELESTIA_BINARY"`
}

var Config EnvConfig

func init() {
	if err := Load(); err != nil {
		panic(err)
	}
}

func Load() error {
	return env.Parse(&Config)
}
