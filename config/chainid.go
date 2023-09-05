package config

import (
	"regexp"
)

var (
	regexChainID         = `[a-z]{1,}`
	regexEIP155Separator = `_{1}`
	regexEIP155          = `[1-9][0-9]*`
	regexEpochSeparator  = `-{1}`
	regexEpoch           = `[1-9][0-9]*`
)

func GetEthID(rollappID string) string {
	re := regexp.MustCompile(`_(\d+)-`)
	matches := re.FindStringSubmatch(rollappID)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
