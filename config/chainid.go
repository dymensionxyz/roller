package config

import (
	"regexp"
)

func GetEthID(rollappID string) string {
	re := regexp.MustCompile(`_(\d+)-`)
	matches := re.FindStringSubmatch(rollappID)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
