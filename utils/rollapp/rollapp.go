package rollapp

import (
	"encoding/json"
	"os/exec"

	"github.com/dymensionxyz/roller/cmd/consts"
	globalutils "github.com/dymensionxyz/roller/cmd/utils"
	tmtypes "github.com/tendermint/tendermint/types"
)

func GetCurrentHeight() (*BlockInformation, error) {
	cmd := getCurrentBlockCmd()
	out, err := globalutils.ExecBashCommandWithStdout(cmd)
	if err != nil {
		return nil, nil
	}

	var blockInfo BlockInformation
	err = json.Unmarshal(out.Bytes(), &blockInfo)
	if err != nil {
		return nil, err
	}

	return &blockInfo, nil
}

func getCurrentBlockCmd() *exec.Cmd {
	cmd := exec.Command(
		consts.Executables.RollappEVM,
		"q",
		"block",
	)
	return cmd
}

type BlockInformation struct {
	BlockId tmtypes.BlockID `json:"block_id"`
	Block   tmtypes.Block   `json:"block"`
}

type Sequencers struct {
	Sequencers []any `json:"sequencers,omitempty"`
}
