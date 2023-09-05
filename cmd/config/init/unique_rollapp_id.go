package initconfig

import (
	"encoding/json"
	"fmt"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"net/http"
	"os/exec"
	"strings"
)

// TODO(#150): roller should use RPC for queries instead of REST
func IsRollappIDUnique(rollappID string, initConfig config.RollappConfig) (bool, error) {
	if initConfig.VMType == config.SDK_ROLLAPP {
		url := initConfig.HubData.API_URL + "/dymensionxyz/dymension/rollapp/rollapp/" + rollappID

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return false, err
		}

		req.Header.Set("accept", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return false, err
		}
		defer resp.Body.Close()

		if resp.StatusCode == 404 {
			return true, nil
		} else if resp.StatusCode == 200 {
			return false, nil
		} else {
			return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
	} else {
		return verifyUniqueEthIdentifier(rollappID, initConfig)
	}
}

type Rollapp struct {
	ID string `json:"rollappId"`
}

type RollappsListResponse struct {
	Rollapps []Rollapp `json:"rollapp"`
}

func verifyUniqueEthIdentifier(rollappID string, rlpCfg config.RollappConfig) (bool, error) {
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
		if config.GetEthID(rollapp.ID) == config.GetEthID(rollappID) {
			return false, nil
		}
	}
	return true, nil
}

func VerifyUniqueRollappID(rollappID string, initConfig config.RollappConfig) error {
	isUniqueRollapp, err := IsRollappIDUnique(rollappID, initConfig)
	if err != nil {
		if initConfig.HubData.ID == consts.Hubs[consts.LocalHubName].ID {
			// When using a local hub and the hub is not yet running, we assume the rollapp ID is unique
			return nil
		}
		return err
	}
	if !isUniqueRollapp {
		if initConfig.VMType == config.SDK_ROLLAPP {
			return fmt.Errorf("rollapp ID \"%s\" already exists on the hub. Please use a unique ID", rollappID)
		} else {
			ethID := config.GetEthID(rollappID)
			return fmt.Errorf("EIP155 ID \"%s\" already exists on the hub (%s). Please use a unique EIP155 ID",
				ethID, strings.Replace(rollappID, ethID, fmt.Sprintf("*%s*", ethID), 1))
		}
	}
	return nil
}
