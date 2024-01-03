package tx_utils

import (
	"bytes"
	"encoding/json"
	"errors"
)

func CheckTxStdOut(stdout bytes.Buffer) error {
	var response Response

	err := json.NewDecoder(&stdout).Decode(&response)
	if err != nil {
		return err
	}

	if response.SDKCode != 0 {
		return errors.New(response.RawLog)
	}
	return nil
}

type Response struct {
	RawLog  string `json:"raw_log"`
	SDKCode int    `json:"code"`
}
