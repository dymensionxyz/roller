package init

import (
	"os"
	"os/exec"
	"path/filepath"

	toml "github.com/pelletier/go-toml"
)

func initializeRollappConfig(rollappExecutablePath string, chainId string, denom string) {
	initRollappCmd := exec.Command(rollappExecutablePath, "init", keyNames.HubSequencer, "--chain-id", chainId, "--home", filepath.Join(os.Getenv("HOME"), configDirName.Rollapp))
	err := initRollappCmd.Run()
	if err != nil {
		panic(err)
	}
	setRollappAppConfig(filepath.Join(os.Getenv("HOME"), configDirName.Rollapp, "config/app.toml"), denom)
}

func setRollappAppConfig(appConfigFilePath string, denom string) {
	config, _ := toml.LoadFile(appConfigFilePath)
	config.Set("minimum-gas-prices", "0"+denom)
	config.Set("api.enable", "true")
	config.Set("grpc.address", "0.0.0.0:8080")
	config.Set("grpc-web.address", "0.0.0.0:8081")
	file, _ := os.Create(appConfigFilePath)
	file.WriteString(config.String())
	file.Close()
}
