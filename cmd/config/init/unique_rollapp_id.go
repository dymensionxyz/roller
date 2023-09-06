package initconfig

import (
	"encoding/json"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"os/exec"
)

type Rollapp struct {
	ID string `json:"rollappId"`
}

type RollappsListResponse struct {
	Rollapps []Rollapp `json:"rollapp"`
}

func isEthIdentifierUnique(ethID string, rlpCfg config.RollappConfig) (bool, error) {
	commonDymdFlags := utils.GetCommonDymdFlags(rlpCfg)
	// TODO: Move the filtering by ethereum rollapp ID logic to the hub
	args := []string{"q", "rollapp", "list", "--limit", "1000000"}
	args = append(args, commonDymdFlags...)
	listRollappCmd := exec.Command(consts.Executables.Dymension, args...)
	out, err := utils.ExecBashCommandWithStdout(listRollappCmd)
	if err != nil {
		return false, err
	}
	rollappsListResponse := RollappsListResponse{}
	err = json.Unmarshal(out.Bytes(), &rollappsListResponse)
	if err != nil {
		return false, err
	}
	for _, rollapp := range rollappsListResponse.Rollapps {
		if config.GetEthID(rollapp.ID) == ethID {
			return false, nil
		}
	}
	return true, nil
}
