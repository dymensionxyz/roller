package sequencer

import (
	"encoding/json"
	"fmt"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/config"
	"io/ioutil"
	"net/http"
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

func getRollappHeight(rollappID string) string {
	resp, err := http.Get(fmt.Sprintf("%s/status", consts.DefaultRollappRPC))
	if err != nil {
		return "-1"
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "-1"
	}
	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return "-1"
	}
	if response.Result.NodeInfo.Network == rollappID {
		return response.Result.SyncInfo.LatestBlockHeight
	} else {
		return "-1"
	}
}

func GetSequencerStatus(cfg config.RollappConfig) string {
	// TODO: Make sure the sequencer status endpoint is being changed after block production is paused.
	height := getRollappHeight(cfg.RollappID)
	if height == "-1" {
		return "Stopped, Restarting..."
	} else {
		return fmt.Sprintf("Active, Height: %s", height)
	}
}
