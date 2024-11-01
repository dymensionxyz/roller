package scripts

import "github.com/dymensionxyz/roller/cmd/consts"

type RaStartupTemplateData struct {
	PasswordFilePath string
	Binary           string
	HomeDir          string
	KeyringBackend   consts.SupportedKeyringBackend
}

type DaStartupTemplateData struct {
	PasswordFilePath string
	Binary           string
	HomeDir          string
	StateNode        string
	KeyringBackend   consts.SupportedKeyringBackend
}
