package keys

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func ParseAddressFromOutput(output *bytes.Buffer) (*KeyInfo, error) {
	key := &KeyInfo{}
	err := json.Unmarshal(output.Bytes(), key)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON output: %v, output: %s", err, output.String())
	}
	return key, nil
}
