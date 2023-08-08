package set

import (
	"errors"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/config"
	"path/filepath"
)

func setLCRPC(cfg config.RollappConfig, value string) error {
	if err := validatePort(value); err != nil {
		return err
	}
	if cfg.DA != config.Celestia {
		return errors.New("setting the LC RPC port is only supported for Celestia")
	}
	return updateFieldInToml(filepath.Join(cfg.Home, consts.ConfigDirName.DALightNode, "config.toml"), "Gateway.Port", value)
}
