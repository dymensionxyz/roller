package healthagent

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/services/load"
	"github.com/dymensionxyz/roller/cmd/services/restart"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/dymint"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/roller"
)

func Start(home string, l *log.Logger) {
	for {
		time.Sleep(15 * time.Second)
		var healthy bool
		localEndpoint := "localhost"
		defaultRaMetricPort := "2112"
		localDaRpcEndpoint := fmt.Sprintf("http://%s:%s", localEndpoint, "26658")

		isDaNodeHealthy, _ := IsEndpointHealthy(localDaRpcEndpoint)
		healthy = isDaNodeHealthy

		submissions, err := QueryPromMetric(
			localEndpoint,
			defaultRaMetricPort,
			"rollapp_consecutive_failed_da_submissions",
		)
		if err != nil {
			l.Println(err)
		}

		if submissions > 10 {
			healthy = false
		}

		// TODO: improve the node swapping, add health checks before swapping etc.
		if !healthy {
			rollerData, err := roller.LoadConfig(home)
			errorhandling.PrettifyErrorIfExists(err)
			rollerConfigPath := roller.GetConfigPath(home)

			i := slices.Index(rollerData.DA.StateNodes, rollerData.DA.CurrentStateNode)
			var newStateNode string
			var nodeIndex int
			if i >= 0 && i+1 < len(rollerData.DA.StateNodes) {
				nodeIndex = i + 1
			} else {
				nodeIndex = 0
			}

			pterm.Warning.Printf(
				"detected problems with DA, hotswapping node to %s\n",
				rollerData.DA.StateNodes[nodeIndex],
			)
			nsn := rollerData.DA.StateNodes[nodeIndex]
			newStateNode = nsn
			err = tomlconfig.UpdateFieldInFile(
				rollerConfigPath,
				"DA.current_state_node",
				newStateNode,
			)
			if err != nil {
				pterm.Error.Println("failed to update state node: ", err)
			}

			servicesToRestart := []string{
				"da-light-client",
			}

			err = load.LoadServices(servicesToRestart, rollerData)
			if err != nil {
				pterm.Error.Println("failed to update services")
			}

			err = restart.RestartSystemdServices(servicesToRestart, home)
			if err != nil {
				pterm.Error.Println("failed to restart services")
			}
		}

		healthy = true
	}
}

func IsEndpointHealthy(url string) (bool, any) {
	// nolint:gosec
	resp, err := http.Get(url)
	if err != nil {
		msg := fmt.Sprintf("Error making request: %v\n", err)
		return false, msg
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		msg := fmt.Sprintf("Error reading response body: %v\n", err)
		return false, msg
	}
	// nolint:errcheck,gosec
	resp.Body.Close()

	var response dymint.RollappHealthResponse
	if json.Valid(body) {
		err = json.Unmarshal(body, &response)
		if err != nil {
			msg := fmt.Sprintf("Error unmarshaling JSON: %v\n", err)
			return false, msg
		}
	} else {
		return false, "invalid json"
	}

	if response.Result.IsHealthy {
		return true, response.Result.Error
	}

	return true, response.Result.Error
}

func QueryPromMetric(host, promMetricPort, metric string) (int, error) {
	endpoint := fmt.Sprintf("http://%s:%s/metrics", host, promMetricPort)
	// nolint: gosec
	resp, err := http.Get(endpoint)
	if err != nil {
		return 0, fmt.Errorf("error fetching metrics: %v", err)
	}
	// nolint: errcheck
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, metric) {
			parts := strings.Fields(line)
			if len(parts) != 2 {
				return 0, fmt.Errorf("unexpected format for metric line: %s", line)
			}
			value, err := strconv.Atoi(parts[1])
			if err != nil {
				return 0, fmt.Errorf("error converting metric value to int: %v", err)
			}
			return value, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("error reading metrics response: %v", err)
	}

	return 0, fmt.Errorf("metric not found: %s", metric)
}
