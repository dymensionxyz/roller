package initconfig

import (
	"os"
	"os/exec"
	"path/filepath"

	toml "github.com/pelletier/go-toml"
)

func initializeRollappConfig(initConfig InitConfig) {
	initRollappCmd := exec.Command(initConfig.RollappBinary, "init", KeyNames.HubSequencer, "--chain-id", initConfig.RollappID, "--home", filepath.Join(initConfig.Home, ConfigDirName.Rollapp))
	err := initRollappCmd.Run()
	if err != nil {
		panic(err)
	}
	setRollappAppConfig(filepath.Join(initConfig.Home, ConfigDirName.Rollapp, "config/app.toml"), initConfig.Denom)
}

func setRollappAppConfig(appConfigFilePath string, denom string) {
	config, _ := toml.LoadFile(appConfigFilePath)
	config.Set("minimum-gas-prices", "0"+denom)
	config.Set("api.enable", "true")
	file, _ := os.Create(appConfigFilePath)
	_, err := file.WriteString(config.String())
	if err != nil {
		panic(err)
	}
	file.Close()
}

func RollappConfigDir(root string) string {
	return filepath.Join(root, ConfigDirName.Rollapp, "config")
}
