package utils

import (
	"bytes"
	"io"
	"net/http"
)

func RestQueryJson(url string) (*bytes.Buffer, error) {
	//nolint:gosec
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	//nolint:errcheck
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(body)
	return buf, nil
}
