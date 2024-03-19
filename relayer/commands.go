package relayer

import (
	"os/exec"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"
)

func (r *Relayer) GetUpdateClientsCmd() *exec.Cmd {
	args := []string{"tx", "update-clients"}
	args = append(args, r.getRelayerDefaultArgs()...)
	return exec.Command(consts.Executables.Relayer, args...)
}

func (r *Relayer) GetRelayAcksCmd() *exec.Cmd {
	args := []string{"tx", "relay-acks"}
	args = append(args, r.getArgsWithSrcChannel()...)
	return exec.Command(consts.Executables.Relayer, args...)
}

func (r *Relayer) GetRelayPacketsCmd() *exec.Cmd {
	args := []string{"tx", "relay-packets"}
	args = append(args, r.getArgsWithSrcChannel()...)
	return exec.Command(consts.Executables.Relayer, args...)
}

// @20240319 the flags `--max-msgs` and `--flush-interval` improve the relayer performance
// a better solution should be implemented as a part of https://github.com/dymensionxyz/roller/issues/769
func (r *Relayer) GetStartCmd() *exec.Cmd {
	args := []string{"start", "--max-msgs", "100", "--flush-interval", "10s"}
	args = append(args, r.getRelayerDefaultArgs()...)
	return exec.Command(consts.Executables.Relayer, args...)
}

func (r *Relayer) getArgsWithSrcChannel() []string {
	return []string{consts.DefaultRelayerPath, r.DstChannel, "--home", filepath.Join(r.Home, consts.ConfigDirName.Relayer)}
}
