package consts

import "path/filepath"

type StartupScriptFilePath string

var StartupScriptFileName = "startup.sh"

var StartupScriptFilePaths = struct {
	RollApp StartupScriptFilePath
	Da      StartupScriptFilePath
}{
	RollApp: StartupScriptFilePath(filepath.Join(ConfigDirName.Rollapp, StartupScriptFileName)),
	Da:      StartupScriptFilePath(filepath.Join(ConfigDirName.DALightNode, StartupScriptFileName)),
}
