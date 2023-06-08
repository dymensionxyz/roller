package initconfig

import (
	"fmt"
	"net/http"
)

func IsRollappIDUnique(rollappID string) (bool, error) {
	url := HubData.API_URL + "/dymensionxyz/dymension/rollapp/rollapp/" + rollappID

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

func VerifyUniqueRollappID(rollappID string) error {
	isUniqueRollapp, err := IsRollappIDUnique(rollappID)
	if err != nil {
		return err
	}
	if !isUniqueRollapp {
		return fmt.Errorf("Rollapp ID \"%s\" already exists on the hub. Please use a unique ID.", rollappID)
	}
	return nil
}
