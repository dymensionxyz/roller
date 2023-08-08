package sequencer

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
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

func getRollappHeight(rlpCfg config.RollappConfig) (string, error) {
	rollappRPCEndpoint, err := GetRPCEndpoint(rlpCfg)
	if err != nil {
		return "-1", err
	}
	resp, err := http.Get(fmt.Sprintf("%s/status", rollappRPCEndpoint))
	if err != nil {
		return "-1", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "-1", err
	}
	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return "-1", err
	}
	if response.Result.NodeInfo.Network == rlpCfg.RollappID {
		return response.Result.SyncInfo.LatestBlockHeight, nil
	} else {
		return "-1", fmt.Errorf("wrong sequencer is running on the machine. Expected network ID %s,"+
			" got %s", rlpCfg.RollappID, response.Result.NodeInfo.Network)
	}
}

type HubResponse struct {
	StateInfo struct {
		StartHeight string `json:"startHeight"`
		NumBlocks   string `json:"numBlocks"`
	} `json:"stateInfo"`
}

func getHubHeight(cfg config.RollappConfig) (string, error) {
	cmd := exec.Command(consts.Executables.Dymension, "q", "rollapp", "state", cfg.RollappID,
		"--output", "json", "--node", cfg.HubData.RPC_URL)
	out, err := utils.ExecBashCommandWithStdout(cmd)
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
func GetSequencerStatus(cfg config.RollappConfig) string {
	// TODO: Make sure the sequencer status endpoint is being changed after block production is paused.
	logger := utils.GetLogger(filepath.Join(cfg.Home, "roller.log"))
	rolHeight, err := getRollappHeight(cfg)
	if err != nil {
		logger.Println(err)
	}
	if rolHeight == "-1" {
		return "Stopped, Restarting..."
	} else {
		hubHeight, err := getHubHeight(cfg)
		if err != nil {
			logger.Println(err)
			return fmt.Sprintf("Active, Height: %s", rolHeight)
		}
		return fmt.Sprintf("Active, Height: %s, Hub: %s", rolHeight, hubHeight)
	}
}
