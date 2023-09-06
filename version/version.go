package version

import "strings"

var (
	BuildVersion = "v0.1.13-rc03-afbhsdbh"
	BuildTime    = "<build-time>"
	BuildCommit  = "<build-commit>"
)

func TrimVersionStr(versionStr string) string {
	return strings.Split(versionStr, "-")[0]
}
