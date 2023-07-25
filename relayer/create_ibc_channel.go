package relayer

import (
	"os/exec"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
)

// Creates an IBC channel between the hub and the client, and return the source channel ID.
func (r *Relayer) CreateIBCChannel(logFileOption utils.CommandOption) (ConnectionChannels, error) {
	//TODO: add support for --override flag
	createClientsCmd := r.getCreateClientsCmd()
	r.logger.Println("Creating clients...")
	if err := utils.ExecBashCmdWithOSOutput(createClientsCmd, logFileOption); err != nil {
		return ConnectionChannels{}, err
	}

	// Before setting up the connection, we need to call update clients
	updateClientsCmd := r.GetUpdateClientsCmd()
	r.logger.Println("Updating clients...")
	if err := utils.ExecBashCmdWithOSOutput(updateClientsCmd, logFileOption); err != nil {
		return ConnectionChannels{}, err
	}

	createConnectionCmd := r.getCreateConnectionCmd()
	r.logger.Println("Creating connection...")
	if err := utils.ExecBashCmdWithOSOutput(createConnectionCmd, logFileOption); err != nil {
		return ConnectionChannels{}, err
	}

	createChannelCmd := r.getCreateChannelCmd()
	r.logger.Println("Creating channel...")
	if err := utils.ExecBashCmdWithOSOutput(createChannelCmd, logFileOption); err != nil {
		return ConnectionChannels{}, err
	}

	r.logger.Println("validating channels...")
	src, dst, err := r.LoadChannels()
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
