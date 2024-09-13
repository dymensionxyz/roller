package eibc

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/docker/docker/client"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	dockerutils "github.com/dymensionxyz/roller/utils/docker"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/pterm/pterm"
)

func GetStartCmd() *exec.Cmd {
	cmd := exec.Command(
		consts.Executables.Eibc,
		"start",
	)
	return cmd
}

func GetInitCmd() *exec.Cmd {
	cmd := exec.Command(
		consts.Executables.Eibc,
		"init",
	)
	return cmd
}

func GetScaleCmd(count string) *exec.Cmd {
	cmd := exec.Command(
		consts.Executables.Eibc,
		"scale",
		count,
	)
	return cmd
}

func GetFundsCmd() *exec.Cmd {
	cmd := exec.Command(
		consts.Executables.Eibc,
		"funds",
	)
	return cmd
}

func GetFulfillOrderCmd(orderId, percentage string, hd consts.HubData) (*exec.Cmd, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(
		consts.Executables.Dymension,
		"tx", "eibc", "fulfill-order",
		orderId, percentage,
		"--from", consts.KeysIds.Eibc,
		"--home", filepath.Join(home, consts.ConfigDirName.Eibc),
		"--fees", fmt.Sprintf("%d%s", consts.DefaultTxFee, consts.Denoms.Hub),
		"--keyring-backend", "test",
		"--node", hd.RPC_URL, "--chain-id", hd.ID,
	)

	return cmd, nil
}

// EnsureWhaleAccount function makes sure that eibc whale account is present in
// the keyring. In eibc client, whale account is the wallet that acts as the bank
// and distributes funds across a set of wallets that fulfill the eibc orders
func EnsureWhaleAccount() error {
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
		addressInfo, err := keys.CreateAddressBinary(kc, home)
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
func CreateMongoDbContainer() error {
	cc, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		fmt.Printf("failed to create docker client: %v\n", err)
		return err
	}

	opts := dockerutils.ContainerConfigOptions{
		Name:   "eibc-mongodb",
		Image:  "mongo:7.0",
		Port:   "27017",
		Envs:   nil,
		Mounts: nil,
	}

	err = dockerutils.CreateContainer(
		context.Background(),
		cc,
		&opts,
	)
	if err != nil {
		fmt.Printf("failed to run mongodb container: %v\n", err)
		return err
	}
	return err
}
