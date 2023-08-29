package relayer

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
)

// CreateIBCChannel Creates an IBC channel between the hub and the client, and return the source channel ID.
func (r *Relayer) CreateIBCChannel(override bool, logFileOption utils.CommandOption) (ConnectionChannels, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	createClientsCmd := r.getCreateClientsCmd(override)
	status := "Creating clients..."
	fmt.Printf("💈 %s\n", status)
	if err := r.WriteRelayerStatus(status); err != nil {
		return ConnectionChannels{}, err
	}
	if err := utils.ExecBashCmd(createClientsCmd, logFileOption); err != nil {
		return ConnectionChannels{}, err
	}

	//after successfull update clients, keep running in the background
	updateClientsCmd := r.GetUpdateClientsCmd()
	utils.RunCommandEvery(ctx, updateClientsCmd.Path, updateClientsCmd.Args[1:], 10, utils.WithDiscardLogging())
	status = "Waiting for block creation..."
	fmt.Printf("💈 %s\n", status)
	if err := r.WriteRelayerStatus(status); err != nil {
		return ConnectionChannels{}, err
	}
	time.Sleep(60 * time.Second)
	createConnectionCmd := r.getCreateConnectionCmd(override)
	status = "Creating connection..."
	fmt.Printf("💈 %s\n", status)
	if err := r.WriteRelayerStatus(status); err != nil {
		return ConnectionChannels{}, err
	}
	if err := utils.ExecBashCmd(createConnectionCmd, logFileOption); err != nil {
		return ConnectionChannels{}, err
	}

	var src, dst string
	createChannelCmd := r.getCreateChannelCmd(override)
	status = "Creating channel..."
	fmt.Printf("💈 %s\n", status)
	if err := r.WriteRelayerStatus(status); err != nil {
		return ConnectionChannels{}, err
	}
	if err := utils.ExecBashCmd(createChannelCmd, logFileOption); err != nil {
		return ConnectionChannels{}, err
	}
	status = "Validating channel established..."
	fmt.Printf("💈 %s\n", status)
	if err := r.WriteRelayerStatus(status); err != nil {
		return ConnectionChannels{}, err
	}

	src, dst, err := r.LoadChannels()
	if err != nil {
		return ConnectionChannels{}, err
	}
	if src == "" || dst == "" {
		return ConnectionChannels{}, fmt.Errorf("could not load channels")
	}

	status = fmt.Sprintf("Active src, %s <-> %s, dst", src, dst)
	if err := r.WriteRelayerStatus(status); err != nil {
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
	args := []string{"tx", "connection", "-t", "300s", "-d"}
	if override {
		args = append(args, "--override")
	}
	args = append(args, r.getRelayerDefaultArgs()...)
	return exec.Command(consts.Executables.Relayer, args...)
}

func (r *Relayer) getCreateChannelCmd(override bool) *exec.Cmd {
	args := []string{"tx", "channel", "-t", "300s", "-r", "5", "-d"}
	if override {
		args = append(args, "--override")
	}
	args = append(args, r.getRelayerDefaultArgs()...)
	return exec.Command(consts.Executables.Relayer, args...)
}

func (r *Relayer) getRelayerDefaultArgs() []string {
	return []string{consts.DefaultRelayerPath, "--home", filepath.Join(r.Home, consts.ConfigDirName.Relayer)}
}
