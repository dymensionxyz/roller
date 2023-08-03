package relayer

import (
	"github.com/dymensionxyz/roller/cmd/consts"
	"os/exec"
	"path/filepath"
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
	args := []string{"tx", "relay-packets", "-l", "1"}
	args = append(args, r.getArgsWithSrcChannel()...)
	return exec.Command(consts.Executables.Relayer, args...)
}

func (r *Relayer) getArgsWithSrcChannel() []string {
	return []string{consts.DefaultRelayerPath, r.SrcChannel, "--home", filepath.Join(r.Home, consts.ConfigDirName.Relayer)}
}
