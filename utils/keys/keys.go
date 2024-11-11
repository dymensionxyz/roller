package keys

import (
	"bytes"
	"encoding/json"
)

func ParseAddressFromOutput(output *bytes.Buffer) (*KeyInfo, error) {
	key := &KeyInfo{}
	err := json.Unmarshal(output.Bytes(), key)
	if err != nil {
		return nil, err
	}
	return key, nil
}
