package set

import (
	"errors"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/config"
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/utils"
	"path/filepath"
)

func setLCGatewayPort(cfg config.RollappConfig, value string) error {
	if err := validatePort(value); err != nil {
		return err
	}
	if cfg.DA != config.Celestia {
		return errors.New("setting the LC RPC port is only supported for Celestia")
	}
	if err := utils.UpdateFieldInToml(filepath.Join(cfg.Home, consts.ConfigDirName.DALightNode, "config.toml"),
		"Gateway.Port", value); err != nil {
		return err
	}
	return sequencer.UpdateDymintDAConfig(cfg)
}

func setLCRPCPort(cfg config.RollappConfig, value string) error {
	if err := validatePort(value); err != nil {
		return err
	}
	if cfg.DA != config.Celestia {
		return errors.New("setting the LC RPC port is only supported for Celestia")
	}
	if err := utils.UpdateFieldInToml(filepath.Join(cfg.Home, consts.ConfigDirName.DALightNode, "config.toml"),
		"RPC.Port", value); err != nil {
		return err
	}
	return sequencer.UpdateDymintDAConfig(cfg)
}
