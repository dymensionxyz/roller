package utils

func GetRelayerDefaultFlags(root string) []string {
	return []string{
		"--src-port", "transfer", "--dst-port", "transfer", "--version", "ics20-1", "--home", root,
	}
}
