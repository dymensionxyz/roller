package eibc

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/docker/client"
	"github.com/pterm/pterm"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
)

// ensureWhaleAccount function makes sure that eibc whale account is present in
// the keyring. In eibc client, whale account is the wallet that acts as the bank
// and distributes funds across a set of wallets that fulfill the eibc orders
func ensureWhaleAccount() error {
	home, _ := os.UserHomeDir()
	kc := utils.KeyConfig{
		Dir:         consts.ConfigDirName.Eibc,
		ID:          consts.KeysIds.Eibc,
		ChainBinary: consts.Executables.Dymension,
		Type:        "",
	}

	_, err := utils.GetAddressInfoBinary(kc, consts.Executables.Dymension)
	if err != nil {
		pterm.Info.Println("whale account not found in the keyring, creating it now")
		addressInfo, err := initconfig.CreateAddressBinary(kc, home)
		if err != nil {
			return err
		}

		addressInfo.Print(utils.WithName(), utils.WithMnemonic())
	}

	return nil
}

// createMongoDbContainer function creates a mongodb container using docker
// sdk. Any 'DOCKER_HOST' can be used for this mongodb container.
// Mongodb is used to store information about processed eibc orders
func createMongoDbContainer() error {
	cc, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		fmt.Printf("failed to create docker client: %v\n", err)
		return err
	}

	cfg := utils.ContainerConfigOptions{
		Name:  "eibc-mongodb",
		Image: "mongo:7.0",
		Port:  "27017",
	}

	err = utils.CreateContainer(
		context.Background(),
		cc,
		&cfg,
	)
	if err != nil {
		fmt.Printf("failed to run mongodb container: %v\n", err)
		return err
	}
	return err
}
