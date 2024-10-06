package set

import (
	"errors"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/roller"
)

func setLCGatewayPort(cfg roller.RollappConfig, value string) error {
	if err := validatePort(value); err != nil {
		return err
	}
	if cfg.DA.Backend != consts.Celestia {
		return errors.New("setting the LC RPC port is only supported for Celestia")
	}
	if err := tomlconfig.UpdateFieldInToml(
		filepath.Join(cfg.Home, consts.ConfigDirName.DALightNode, "config.toml"),
		"Gateway.Port", value,
	); err != nil {
		return err
	}
	return sequencer.UpdateDymintDAConfig(cfg)
}

func setLCRPCPort(cfg roller.RollappConfig, value string) error {
	if err := validatePort(value); err != nil {
		return err
	}
	if cfg.DA.Backend != consts.Celestia {
		return errors.New("setting the LC RPC port is only supported for Celestia")
	}
	if err := tomlconfig.UpdateFieldInToml(
		filepath.Join(cfg.Home, consts.ConfigDirName.DALightNode, "config.toml"),
		"RPC.Port", value,
	); err != nil {
		return err
	}
	return sequencer.UpdateDymintDAConfig(cfg)
}
