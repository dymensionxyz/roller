package utils

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

func RestQueryJson(url string) (*bytes.Buffer, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(body)
	return buf, nil
}
