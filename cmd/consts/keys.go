package consts

type SupportedKeyringBackend string

var SupportedKeyringBackends = struct {
	OS   SupportedKeyringBackend
	Test SupportedKeyringBackend
}{
	OS:   "os",
	Test: "test",
}

type OsKeyringPwdFileName string

var OsKeyringPwdFileNames = struct {
	RollApp OsKeyringPwdFileName
	Da      OsKeyringPwdFileName
}{
	RollApp: ".ra-os-keyring-psw",
	Da:      ".da-os-keyring-psw",
}
