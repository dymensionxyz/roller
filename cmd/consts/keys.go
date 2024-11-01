package consts

import (
	"path/filepath"
)

type SupportedKeyringBackend string

var SupportedKeyringBackends = struct {
	OS   SupportedKeyringBackend
	Test SupportedKeyringBackend
}{
	OS:   "os",
	Test: "test",
}

type OsKeyringPwdFilePath string

// nolint: gosec
var OsKeyringPwdFileName = ".os-keyring-psw"

var OsKeyringPwdFilePaths = struct {
	RollApp OsKeyringPwdFilePath
	Da      OsKeyringPwdFilePath
}{
	RollApp: OsKeyringPwdFilePath(filepath.Join(ConfigDirName.Rollapp, OsKeyringPwdFileName)),
	Da:      OsKeyringPwdFilePath(filepath.Join(ConfigDirName.DALightNode, OsKeyringPwdFileName)),
}
