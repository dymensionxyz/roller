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
func (r *Relayer) CreateIBCChannel(logFileOption utils.CommandOption) (ConnectionChannels, error) {
	//TODO: add support for --override flag
	createClientsCmd := r.getCreateClientsCmd()
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

	createConnectionCmd := r.getCreateConnectionCmd()
	fmt.Println("💈 Creating connection...")
	if err := utils.ExecBashCmd(createConnectionCmd, logFileOption); err != nil {
		return ConnectionChannels{}, err
	}

	createChannelCmd := r.getCreateChannelCmd()
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

func (r *Relayer) getCreateClientsCmd() *exec.Cmd {
	args := []string{"tx", "clients"}
	args = append(args, r.getRelayerDefaultArgs()...)
	return exec.Command(consts.Executables.Relayer, args...)
}

func (r *Relayer) getCreateConnectionCmd() *exec.Cmd {
	args := []string{"tx", "connection", "-t", "300s"}
	args = append(args, r.getRelayerDefaultArgs()...)
	return exec.Command(consts.Executables.Relayer, args...)
}

func (r *Relayer) getCreateChannelCmd() *exec.Cmd {
	args := []string{"tx", "channel", "-t", "300s", "--override"}
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
