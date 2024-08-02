package initrollapp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/config"
)

type DenomUnits struct {
	Denom    string `json:"denom"`
	Exponent int    `json:"exponent"`
}

type DenomMetadata struct {
	Description string       `json:"description"`
	DenomUnits  []DenomUnits `json:"denom_units"`
	Base        string       `json:"base"`
	Display     string       `json:"display"`
	Name        string       `json:"name"`
	Symbol      string       `json:"symbol"`
}

func NewDenomMetadata(raCfg config.RollappConfig) *[]DenomMetadata {
	du := []DenomUnits{
		{
			Denom:    raCfg.BaseDenom,
			Exponent: 0,
		},
		{
			Denom:    raCfg.Denom,
			Exponent: int(raCfg.Decimals),
		},
	}

	dm := DenomMetadata{
		Description: fmt.Sprintf(
			"The native staking and governance token of the %s",
			raCfg.RollappID,
		),
		DenomUnits: du,
		Base:       raCfg.BaseDenom,
		Display:    raCfg.Denom,
		Name:       raCfg.Denom,
		Symbol:     raCfg.Denom,
	}

	res := []DenomMetadata{dm}

	return &res
}

func WriteDenomMetadata(raCfg config.RollappConfig) error {
	path := filepath.Join(
		raCfg.Home,
		consts.ConfigDirName.Rollapp,
		"init",
		"denom-metadata.json",
	)

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		return err
	}

	dm := NewDenomMetadata(raCfg)
	j, _ := json.Marshal(dm)

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	// nolint errcheck
	defer f.Close()

	_, err = f.Write(j)
	if err != nil {
		return err
	}

	_, err = os.Stat(path)
	if err != nil {
		return err
	}

	return nil
}
