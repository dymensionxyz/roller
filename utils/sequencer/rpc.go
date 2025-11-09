package sequencer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type tendermintStatusResponse struct {
	Result struct {
		NodeInfo struct {
			Network string `json:"network"`
		} `json:"node_info"`
	} `json:"result"`
}

func ValidateRollappRPCEndpoint(rollappID, endpoint string) (string, error) {
	trimmed := strings.TrimSpace(endpoint)
	if trimmed == "" {
		return "", fmt.Errorf("rollapp rpc endpoint is empty")
	}

	normalized := normalizeRpcScheme(trimmed)
	normalized = strings.TrimRight(normalized, "/")

	statusURL := normalized + "/status"
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(statusURL)
	if err != nil {
		return "", fmt.Errorf("failed to reach rollapp rpc endpoint %s: %w", normalized, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf(
			"rollapp rpc endpoint %s returned %s from /status",
			normalized,
			resp.Status,
		)
	}

	var status tendermintStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return "", fmt.Errorf("failed to decode /status from %s: %w", statusURL, err)
	}

	chainID := strings.TrimSpace(status.Result.NodeInfo.Network)
	if chainID == "" {
		return "", fmt.Errorf("rollapp rpc endpoint %s returned empty chain-id in /status response", normalized)
	}
	if chainID != rollappID {
		return "", fmt.Errorf(
			"rollapp rpc endpoint %s reports chain-id %s, expected %s",
			normalized,
			chainID,
			rollappID,
		)
	}

	return normalized, nil
}

func normalizeRpcScheme(endpoint string) string {
	switch {
	case strings.HasPrefix(endpoint, "tcp://"):
		return "http://" + strings.TrimPrefix(endpoint, "tcp://")
	case strings.HasPrefix(endpoint, "ws://"):
		return "http://" + strings.TrimPrefix(endpoint, "ws://")
	case strings.HasPrefix(endpoint, "wss://"):
		return "https://" + strings.TrimPrefix(endpoint, "wss://")
	case strings.HasPrefix(endpoint, "http://"), strings.HasPrefix(endpoint, "https://"):
		return endpoint
	default:
		return "https://" + endpoint
	}
}
