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

type Result struct {
	NodeInfo NodeInfo `json:"node_info"`
}

type Response struct {
	Result Result `json:"result"`
}

func IsSeqListen(rollappID string) bool {
	resp, err := http.Get(fmt.Sprintf("%s/status", consts.DefaultRollappRPC))
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return false
	}
	return response.Result.NodeInfo.Network == rollappID
}

func GetSequencerStatus(cfg config.RollappConfig) string {
	// TODO: Make sure the sequencer status endpoint is being changed after block production is paused.
	if IsSeqListen(cfg.RollappID) {
		return "Active"
	}
	return "Stopped, Restarting..."
}
