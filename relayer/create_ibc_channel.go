package relayer

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/sequencer"
)

// CreateIBCChannel Creates an IBC channel between the hub and the client,
// and return the source channel ID.
func (r *Relayer) CreateIBCChannel(
	override bool,
	logFileOption utils.CommandOption,
	seq *sequencer.Sequencer,
) (ConnectionChannels, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run send funds command from sequencer to itself to make sure the chain is
	// progressing for connection and channel creation.
	// replaced update clients to avoid account sequence mismatch and
	// premature heights updates e.g "TrustedHeight {1 x} must be less than header height {1 y}"
	sequencerAddress, err := utils.GetAddressBinary(
		utils.KeyConfig{
			Dir: filepath.Join(seq.RlpCfg.Home, consts.ConfigDirName.Rollapp),
			ID:  consts.KeysIds.RollappSequencer,
		}, consts.Executables.RollappEVM,
	)
	if err != nil {
		return ConnectionChannels{}, err
	}

	sendFundsCmd := seq.GetSendCmd(sequencerAddress)
	utils.RunCommandEvery(
		ctx,
		sendFundsCmd.Path,
		sendFundsCmd.Args[1:],
		5,
		utils.WithDiscardLogging(),
	)

	var status string

	// Create client if it doesn't exist or override is true
	clientsExist := false
	if !override {
		// Check if clients exist
		clientsExist, _ = r.CheckClientsExist()
	}
	if !clientsExist {
		// wait for block to be created
		pterm.Info.Println("💈 Validating rollapp height > 2 before creating clients...")
		if err := r.WriteRelayerStatus(status); err != nil {
			return ConnectionChannels{}, err
		}

		if err := waitForValidRollappHeight(seq); err != nil {
			fmt.Println(err)
			return ConnectionChannels{}, err
		}

		// We always pass override otherwise this command hangs if there are too many clients
		createClientsCmd := r.getCreateClientsCmd(true)
		pterm.Info.Println("💈 Creating clients...")
		if err := r.WriteRelayerStatus(status); err != nil {
			return ConnectionChannels{}, err
		}

		if err := utils.ExecBashCmd(createClientsCmd, logFileOption); err != nil {
			fmt.Println(err)
			return ConnectionChannels{}, err
		}
	}

	// Sleep for a few seconds to make sure the clients are created
	// otherwise the connection creation attempt fails
	time.Sleep(10 * time.Second)

	connectionID, _ := r.GetActiveConnection()
	if connectionID == "" || override {
		pterm.Info.Println("💈 Creating connection...")
		if err := r.WriteRelayerStatus(status); err != nil {
			return ConnectionChannels{}, err
		}
		createConnectionCmd := r.getCreateConnectionCmd(override)
		if err := utils.ExecBashCmd(createConnectionCmd, logFileOption); err != nil {
			return ConnectionChannels{}, err
		}
	}

	var src, dst string

	// Sleep for a few seconds to make sure the connection is created
	time.Sleep(15 * time.Second)
	// we ran create channel with override, as it not recovarable anyway
	createChannelCmd := r.getCreateChannelCmd(true)
	pterm.Info.Println("💈 Creating channel...")
	if err := r.WriteRelayerStatus(status); err != nil {
		return ConnectionChannels{}, err
	}
	if err := utils.ExecBashCmd(createChannelCmd, logFileOption); err != nil {
		return ConnectionChannels{}, err
	}
	status = ""
	pterm.Info.Println("💈 Validating channel established...")
	if err := r.WriteRelayerStatus(status); err != nil {
		return ConnectionChannels{}, err
	}

	_, _, err = r.LoadActiveChannel()
	if err != nil {
		return ConnectionChannels{}, err
	}
	if r.SrcChannel == "" || r.DstChannel == "" {
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

// waitForValidRollappHeight waits for the rollapp height to be greater than 2 otherwise
// it will fail to create clients.
func waitForValidRollappHeight(seq *sequencer.Sequencer) error {
	logger := utils.GetRollerLogger(seq.RlpCfg.Home)
	for {
		rollappHeightStr, err := seq.GetRollappHeight()
		if err != nil {
			logger.Printf("💈 Error getting rollapp height, %s", err.Error())
			continue
		}
		rollappHeight, err := strconv.Atoi(rollappHeightStr)
		if err != nil {
			logger.Printf("💈 Error converting rollapp height to int, %s", err.Error())
			continue
		}
		if rollappHeight <= 2 {
			logger.Printf("💈 Waiting for rollapp height to be greater than 2")
			continue
		}
		return nil
	}
}

func (r *Relayer) getCreateClientsCmd(override bool) *exec.Cmd {
	args := []string{"tx", "clients"}
	args = append(args, r.getRelayerDefaultArgs()...)
	args = append(args, "--log-level", "debug")
	if override {
		args = append(args, "--override")
	}
	cmd := exec.Command(consts.Executables.Relayer, args...)
	return cmd
}

func (r *Relayer) getCreateConnectionCmd(override bool) *exec.Cmd {
	args := []string{"tx", "connection"}
	if override {
		args = append(args, "--override")
	}
	args = append(args, r.getRelayerDefaultArgs()...)
	return exec.Command(consts.Executables.Relayer, args...)
}

func (r *Relayer) getCreateChannelCmd(override bool) *exec.Cmd {
	args := []string{"tx", "channel", "-t", "60s", "-d"}
	if override {
		args = append(args, "--override")
	}
	args = append(args, r.getRelayerDefaultArgs()...)
	return exec.Command(consts.Executables.Relayer, args...)
}

func (r *Relayer) getRelayerDefaultArgs() []string {
	return []string{
		consts.DefaultRelayerPath,
		"--home",
		filepath.Join(r.Home, consts.ConfigDirName.Relayer),
	}
}
