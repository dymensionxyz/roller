package eibc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/docker/client"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
)

func ensureWhaleAccount() error {
	home, _ := os.UserHomeDir()
	eibcHome := filepath.Join(home, ".order-client")
	kc := utils.KeyConfig{
		Dir:         eibcHome,
		ID:          "client",
		ChainBinary: consts.Executables.Dymension,
		Type:        "",
	}

	_, err := utils.GetAddressBinary(kc, consts.Executables.Dymension)
	if err != nil {
		fmt.Println("whale account not found in the keyring, creating it now")
		addressInfo, err := initconfig.CreateAddressBinaryWithSensitiveOutput(kc, home)
		if err != nil {
			return err
		}

		whaleAddress := utils.SecretAddressData{
			AddressData: utils.AddressData{
				Name: addressInfo.Name,
				Addr: addressInfo.Address,
			},
			Mnemonic: addressInfo.Mnemonic,
		}

		utils.PrintSecretAddressesWithTitle([]utils.SecretAddressData{whaleAddress})
	}

	return nil
}

func createMongoDbContainer() error {
	cc, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		fmt.Printf("failed to create docker client: %v\n", err)
		return err
	}

	err = utils.CheckAndCreateMongoDBContainer(
		context.Background(),
		cc,
	)
	if err != nil {
		fmt.Printf("failed to run mongodb container: %v\n", err)
		return err
	}
	return err
}
