package initconfig

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/dymensionxyz/roller/config"
)

// TODO(#150): roller should use RPC for queries instead of REST
func IsRollappIDUnique(rollappID string, initConfig config.RollappConfig) (bool, error) {
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
}

func VerifyUniqueRollappID(rollappID string, initConfig config.RollappConfig) error {
	isUniqueRollapp, err := IsRollappIDUnique(rollappID, initConfig)
	if err != nil {
		_, ok := err.(*url.Error)
		if ok && initConfig.HubData.ID == LocalHubID {
			// When using a local hub and the hub is not yet running, we assume the rollapp ID is unique
			return nil
		}
		return err
	}
	if !isUniqueRollapp {
		return fmt.Errorf("Rollapp ID \"%s\" already exists on the hub. Please use a unique ID.", rollappID)
	}
	return nil
}
