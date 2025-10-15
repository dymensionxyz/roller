package sequencer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strconv"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
)

type NodeInfo struct {
	Network string `json:"network"`
}

type SyncInfo struct {
	LatestBlockHeight string `json:"latest_block_height"`
}

type Result struct {
	NodeInfo NodeInfo `json:"node_info"`
	SyncInfo SyncInfo `json:"sync_info"`
}

type Response struct {
	Result Result `json:"result"`
}

type HubResponse struct {
	StateInfo struct {
		StartHeight    string `json:"startHeight"`
		NumBlocks      string `json:"numBlocks"`
		CreationHeight string `json:"creationHeight"`
	} `json:"stateInfo"`
}

type HealthResult struct {
	IsHealthy bool   `json:"isHealthy"`
	Error     string `json:"error"`
}

type HealthResponse struct {
	JsonRPC string       `json:"jsonrpc"`
	Result  HealthResult `json:"result"`
	ID      int          `json:"id"`
}

func (seq *Sequencer) GetRollappHeight() (string, error) {
	rollappRPCEndpoint := seq.GetRPCEndpoint()
	resp, err := http.Get(fmt.Sprintf("%s/status", rollappRPCEndpoint))
	if err != nil {
		return "-1", err
	}

	//nolint:errcheck
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "-1", err
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return "-2", err
	}
	if response.Result.NodeInfo.Network == seq.RlpCfg.RollappID {
		return response.Result.SyncInfo.LatestBlockHeight, nil
	} else {
		return "-1", fmt.Errorf(
			"wrong sequencer is running on the machine. Expected network ID %s,"+
				" got %s", seq.RlpCfg.RollappID, response.Result.NodeInfo.Network,
		)
	}
}

func GetFirstStateUpdateHeight(raID, hubRpcEndpoint, hubChainID string) (int, error) {
	cmd := exec.Command(
		consts.Executables.Dymension,
		"q",
		"rollapp",
		"state",
		"--index",
		"1",
		raID,
		"--output",
		"json",
		"--node",
		hubRpcEndpoint,
		"--chain-id",
		hubChainID,
	)

	out, err := bash.ExecCommandWithStdoutFiltered(cmd)
	if err != nil {
		return 0, err
	}

	var resp HubResponse
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		return 0, err
	}
	h, err := strconv.Atoi(resp.StateInfo.CreationHeight)
	if err != nil {
		return 0, fmt.Errorf("unable to convert start height to int: %s", err)
	}

	return h, nil
}

func (seq *Sequencer) GetHubHeight() (string, error) {
	cmd := exec.Command(
		consts.Executables.Dymension,
		"q",
		"rollapp",
		"state",
		seq.RlpCfg.RollappID,
		"--output",
		"json",
		"--node",
		seq.RlpCfg.HubData.RpcUrl,
		"--chain-id",
		seq.RlpCfg.HubData.ID,
	)

	out, err := bash.ExecCommandWithStdoutFiltered(cmd)
	if err != nil {
		return "", err
	}

	var resp HubResponse
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		return "", err
	}

	startHeight, err := strconv.Atoi(resp.StateInfo.StartHeight)
	if err != nil {
		return "", fmt.Errorf("unable to convert start height to int: %s", err)
	}
	numBlocks, err := strconv.Atoi(resp.StateInfo.NumBlocks)
	if err != nil {
		return "", fmt.Errorf("unable to convert num blocks to int: %s", err)
	}
	return strconv.Itoa(startHeight + numBlocks - 1), nil
}

func (seq *Sequencer) GetSequencerHealth() error {
	var res HealthResponse
	url := seq.GetLocalEndpoint(seq.RPCPort)
	healthEndpoint := fmt.Sprintf("%s/health", url)

	// nolint gosec
	resp, err := http.Get(healthEndpoint)
	if err != nil {
		return err
	}

	// nolint errcheck
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &res)
	if err != nil {
		return err
	}

	if !res.Result.IsHealthy {
		return errors.New(res.Result.Error)
	}

	return nil
}

func (seq *Sequencer) GetSequencerStatus() string {
	// TODO: Make sure the sequencer status endpoint is being changed after block production is paused.
	rolHeight, err := seq.GetRollappHeight()
	if err != nil {
		seq.logger.Println(err)
	}

	// ?
	if rolHeight == "-1" {
		return "Stopped, Restarting..."
	}

	err = seq.GetSequencerHealth()
	if err != nil {
		return fmt.Sprintf(
			`
status: Unhealthy
error: %v
`, err,
		)
	}

	hubHeight, err := seq.GetHubHeight()
	if err != nil {
		seq.logger.Println(err)

		err := seq.ReadPorts()
		if err != nil {
			fmt.Println("failed to retrieve ports: ", err)
		}

		return fmt.Sprintf(
			`RollApp
status: Healthy
height: %s
`, rolHeight,
		)
	}

	return fmt.Sprintf(
		`RollApp:
status: Healthy
height: %s

Hub:
height: %s`, rolHeight, hubHeight,
	)
}
