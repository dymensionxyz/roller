package tx_utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
)

func CheckTxStdOut(stdout bytes.Buffer) error {
	var response Response

	err := json.NewDecoder(&stdout).Decode(&response)
	if err != nil {
		return err
	}

	if strings.Contains(response.RawLog, "fail") || strings.Contains(response.RawLog, "error") ||
		strings.Contains(response.RawLog, "insufficient funds") {
		return errors.New(response.RawLog)
	}
	return nil
}

type Response struct {
	RawLog string `json:"raw_log"`
}
