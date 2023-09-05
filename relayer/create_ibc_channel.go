package relayer

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/dymensionxyz/roller/sequencer"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
)

// CreateIBCChannel Creates an IBC channel between the hub and the client, and return the source channel ID.
func (r *Relayer) CreateIBCChannel(override bool, logFileOption utils.CommandOption, seq *sequencer.Sequencer,
) (ConnectionChannels, error) {
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

	//after successful update clients, keep running in the background
	updateClientsCmd := r.GetUpdateClientsCmd()
	utils.RunCommandEvery(ctx, updateClientsCmd.Path, updateClientsCmd.Args[1:], 10, utils.WithDiscardLogging())

	//wait for block to be created
	status = "Creating block..."
	fmt.Printf("💈 %s\n", status)
	if err := r.WriteRelayerStatus(status); err != nil {
		return ConnectionChannels{}, err
	}
	if err := waitForValidRollappHeight(seq); err != nil {
		return ConnectionChannels{}, err
	}

	connectionID, _ := r.GetActiveConnection()
	if connectionID == "" {
		status = "Creating connection..."
		fmt.Printf("💈 %s\n", status)
		if err := r.WriteRelayerStatus(status); err != nil {
			return ConnectionChannels{}, err
		}
		createConnectionCmd := r.getCreateConnectionCmd(override)
		if err := utils.ExecBashCmd(createConnectionCmd, logFileOption); err != nil {
			return ConnectionChannels{}, err
		}
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

	src, dst, err := r.LoadActiveChannel()
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

func waitForValidRollappHeight(seq *sequencer.Sequencer) error {
	initialHubHeightStr, err := seq.GetHubHeight()
	if err != nil {
		return err
	}
	initialHubHeight, err := strconv.Atoi(initialHubHeightStr)
	if err != nil {
		return err
	}
	initialRollappHeightStr, err := seq.GetRollappHeight()
	if err != nil {
		return err
	}
	initialRollappHeight, err := strconv.Atoi(initialRollappHeightStr)
	if err != nil {
		return err
	}
	for {
		time.Sleep(30 * time.Second)
		hubHeightStr, err := seq.GetHubHeight()
		if err != nil {
			fmt.Printf("💈 Error getting rollapp hub height, %s", err.Error())
			continue
		}
		hubHeight, err := strconv.Atoi(hubHeightStr)
		if err != nil {
			fmt.Printf("💈 Error converting hub height to int, %s", err.Error())
			continue
		}
		if hubHeight < 3 {
			fmt.Printf("💈 Waiting for hub height to be greater than 2, current height: %d\n", hubHeight)
			continue
		}
		if hubHeight <= initialHubHeight {
			fmt.Printf("💈 Waiting for hub height to be greater than initial height,"+
				" initial height: %d,current height: %d\n", initialHubHeight, hubHeight)
			continue
		}
		rollappHeightStr, err := seq.GetRollappHeight()
		if err != nil {
			fmt.Printf("💈 Error getting rollapp height, %s", err.Error())
			continue
		}
		rollappHeight, err := strconv.Atoi(rollappHeightStr)
		if err != nil {
			fmt.Printf("💈 Error converting rollapp height to int, %s", err.Error())
			continue
		}
		if rollappHeight <= initialRollappHeight {
			fmt.Printf("💈 Waiting for rollapp height to be greater than initial height,"+
				" initial height: %d,current height: %d\n", initialRollappHeight, rollappHeight)
		}
		return nil
	}
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
