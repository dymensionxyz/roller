package tx_utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
)

func CheckTxJsonStdOut(stdout bytes.Buffer) error {
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

func CheckTxYamlStdOut(stdout bytes.Buffer) error {
	var response Response
	err := yaml.Unmarshal(stdout.Bytes(), &response)
	if err != nil {
		fmt.Printf("Error parsing YAML: %v\n", err)
		return err
	}

	if response.SDKCode != 0 {
		return errors.New(response.RawLog)
	}

	return nil
}

type Response struct {
	RawLog  string `json:"raw_log" yaml:"raw_log"`
	SDKCode int    `json:"code"    yaml:"code"`
}
