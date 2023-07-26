package relayer

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	retry "github.com/avast/retry-go"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
)

// Creates an IBC channel between the hub and the client, and return the source channel ID.
func (r *Relayer) CreateIBCChannel(override bool, logFileOption utils.CommandOption) (ConnectionChannels, error) {
	createClientsCmd := r.getCreateClientsCmd(override)
	fmt.Println("💈 Creating clients...")
	if err := utils.ExecBashCmd(createClientsCmd, logFileOption); err != nil {
		return ConnectionChannels{}, err
	}

	// Before setting up the connection, we need to call update clients

	fmt.Println("💈 Updating clients...")
	err := retry.Do(
		func() error {
			updateClientsCmd := r.GetUpdateClientsCmd()
			return utils.ExecBashCmd(updateClientsCmd, logFileOption)
		},
		retry.Delay(time.Duration(10)*time.Second),
		retry.DelayType(retry.FixedDelay),
		retry.Attempts(5),
		retry.OnRetry(func(n uint, err error) {
			r.logger.Printf("error updating clients. attempt %d, error %s", n, err)
		}),
	)
	if err != nil {
		return ConnectionChannels{}, err
	}

	createConnectionCmd := r.getCreateConnectionCmd(override)
	fmt.Println("💈 Creating connection...")
	if err := utils.ExecBashCmd(createConnectionCmd, logFileOption); err != nil {
		return ConnectionChannels{}, err
	}

	createChannelCmd := r.getCreateChannelCmd(override)
	fmt.Println("💈 Creating channel...")
	if err := utils.ExecBashCmd(createChannelCmd, logFileOption); err != nil {
		return ConnectionChannels{}, err
	}

	fmt.Println("💈 Validating channels...")
	var src, dst string
	err = retry.Do(
		func() error {
			var err error
			src, dst, err = r.LoadChannels()
			if err != nil {
				return err
			}
			if src == "" || dst == "" {
				return fmt.Errorf("could not load channels")
			}
			return nil
		},
		retry.Delay(time.Duration(10)*time.Second),
		retry.DelayType(retry.FixedDelay),
		retry.Attempts(5),
		retry.OnRetry(func(n uint, err error) {
			r.logger.Printf("error updating clients. attempt %d, error %s", n, err)
		}),
	)
	if err != nil {
		return ConnectionChannels{}, err
	}

	return ConnectionChannels{
		Src: src,
		Dst: dst,
	}, nil
}

func (r *Relayer) getCreateClientsCmd(override bool) *exec.Cmd {
	args := []string{"tx", "clients"}
	args = append(args, r.getRelayerDefaultArgs()...)
	if override {
		args = append(args, "--override")
	}
	return exec.Command(consts.Executables.Relayer, args...)
}

func (r *Relayer) getCreateConnectionCmd(override bool) *exec.Cmd {
	args := []string{"tx", "connection", "-t", "300s"}
	if override {
		args = append(args, "--override")
	}
	args = append(args, r.getRelayerDefaultArgs()...)
	return exec.Command(consts.Executables.Relayer, args...)
}

func (r *Relayer) getCreateChannelCmd(override bool) *exec.Cmd {
	args := []string{"tx", "channel", "-t", "300s", "--override"}
	if override {
		args = append(args, "--override")
	}
	args = append(args, r.getRelayerDefaultArgs()...)
	return exec.Command(consts.Executables.Relayer, args...)
}

func (r *Relayer) GetUpdateClientsCmd() *exec.Cmd {
	args := []string{"tx", "update-clients"}
	args = append(args, r.getRelayerDefaultArgs()...)
	return exec.Command(consts.Executables.Relayer, args...)
}

func (r *Relayer) getRelayerDefaultArgs() []string {
	return []string{consts.DefaultRelayerPath, "--home", filepath.Join(r.Home, consts.ConfigDirName.Relayer)}
}
