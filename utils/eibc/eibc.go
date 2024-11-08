package eibc

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/docker/docker/client"
	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/config/yamlconfig"
	dockerutils "github.com/dymensionxyz/roller/utils/docker"
	"github.com/dymensionxyz/roller/utils/keys"
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

func GetFulfillOrderCmd(orderId, fee string, hd consts.HubData) (*exec.Cmd, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(
		consts.Executables.Dymension,
		"tx", "eibc", "fulfill-order",
		orderId, fee,
		"--from", consts.KeysIds.Eibc,
		"--home", filepath.Join(home, consts.ConfigDirName.Eibc),
		"--fees", fmt.Sprintf("%d%s", consts.DefaultTxFee, consts.Denoms.Hub),
		"--keyring-backend", string(consts.SupportedKeyringBackends.Test),
		"--node", hd.RpcUrl, "--chain-id", hd.ID,
	)

	return cmd, nil
}

// EnsureWhaleAccount function makes sure that eibc whale account is present in
// the keyring. In eibc client, whale account is the wallet that acts as the bank
// and distributes funds across a set of wallets that fulfill the eibc orders
func EnsureWhaleAccount() error {
	home, _ := os.UserHomeDir()
	kc := keys.KeyConfig{
		Dir:            consts.ConfigDirName.Eibc,
		ID:             consts.KeysIds.Eibc,
		ChainBinary:    consts.Executables.Dymension,
		Type:           "",
		KeyringBackend: consts.SupportedKeyringBackends.Test,
	}

	_, err := kc.Info(home)
	if err != nil {
		pterm.Info.Println("whale account not found in the keyring, creating it now")
		addressInfo, err := kc.Create(home)
		if err != nil {
			return err
		}

		addressInfo.Print(keys.WithName(), keys.WithMnemonic())
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

func AddRollappToEibc(value, rollAppID, eibcHome string) error {
	eibcConfigPath := filepath.Join(eibcHome, "config.yaml")

	vf, _, err := big.ParseFloat(value, 10, 64, big.ToNearestEven)
	valueFloat, _ := vf.Float32()
	if err != nil {
		return fmt.Errorf("failed to convert value to float: %v", err)
	}

	updates := map[string]interface{}{
		fmt.Sprintf("fulfill_criteria.min_fee_percentage.chain.%s", rollAppID): valueFloat,
	}
	err = yamlconfig.UpdateNestedYAML(eibcConfigPath, updates)
	if err != nil {
		return fmt.Errorf("failed to update config: %v", err)
	}
	return nil
}
