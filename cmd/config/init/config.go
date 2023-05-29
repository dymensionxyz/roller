package init

import (
	"os"

	toml "github.com/pelletier/go-toml"
)

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