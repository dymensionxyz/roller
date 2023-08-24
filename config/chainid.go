package config

import (
	"fmt"
	"regexp"
)

var (
	regexChainID         = `[a-z]{1,}`
	regexEIP155Separator = `_{1}`
	regexEIP155          = `[1-9][0-9]*`
	regexEpochSeparator  = `-{1}`
	regexEpoch           = `[1-9][0-9]*`
	rollappID            = regexp.MustCompile(fmt.Sprintf(`^(%s)%s(%s)%s(%s)$`,
		regexChainID,
		regexEIP155Separator,
		regexEIP155,
		regexEpochSeparator,
		regexEpoch))
)

func ValidateRollAppID(id string) error {
	if !rollappID.MatchString(id) {
		return fmt.Errorf("invalid RollApp ID '%s'. "+
			"Expected format: name_uniqueID-revision (e.g. myrollapp_1234-1)", id)
	}
	return nil
}

func GetEthID(rollappID string) string {
	re := regexp.MustCompile(`_(\d+)-`)
	matches := re.FindStringSubmatch(rollappID)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
