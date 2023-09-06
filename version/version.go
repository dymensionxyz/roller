package version

import "strings"

var (
	BuildVersion = "<version>"
	BuildTime    = "<build-time>"
	BuildCommit  = "<build-commit>"
)

func TrimVersionStr(versionStr string) string {
	return strings.Split(versionStr, "-")[0]
}
