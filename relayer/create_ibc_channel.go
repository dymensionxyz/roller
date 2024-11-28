package relayer

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/logging"
)

// CreateIBCChannel Creates an IBC channel between the hub and the client,
// and return the source channel ID.
func (r *Relayer) CreateIBCChannel(
	logFileOption bash.CommandOption,
	raData consts.RollappData,
	hd consts.HubData,
) (ConnectionChannels, error) {
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	var status string

	// TODO: this is probably not true anymore, review and remove the sleep if necessary
	// Sleep for a few seconds to make sure the clients are created
	// otherwise the connection creation attempt fails
	time.Sleep(10 * time.Second)

	connectionID, _, err := r.GetActiveConnectionIDs(raData, hd)
	if err != nil {
		return ConnectionChannels{}, err
	}

	if connectionID == "" {
		pterm.Info.Println("💈 Creating connection...")
		if err := r.WriteRelayerStatus(status); err != nil {
			return ConnectionChannels{}, err
		}

		sp, err := getHubStakingParams(r.Hub)
		if err != nil {
			return ConnectionChannels{}, err
		}

		createConnectionCmd := r.getCreateConnectionCmd(sp.UnbondingTime)
		if err := bash.ExecCmd(createConnectionCmd, logFileOption); err != nil {
			return ConnectionChannels{}, err
		}
	}

	// Sleep for a few seconds to make sure the connection is created
	time.Sleep(15 * time.Second)
	// we ran create channel with override, as it not recovarable anyway
	createChannelCmd := r.getCreateChannelCmd(false)

	pterm.Info.Println("💈 Creating channel (this may take a while)...")
	if err := r.WriteRelayerStatus(status); err != nil {
		return ConnectionChannels{}, err
	}
	if err := bash.ExecCmd(createChannelCmd, logFileOption); err != nil {
		return ConnectionChannels{}, err
	}
	status = ""
	pterm.Info.Println("💈 Validating channel established...")
	if err := r.WriteRelayerStatus(status); err != nil {
		return ConnectionChannels{}, err
	}

	err = r.LoadActiveChannel(raData, hd)
	if err != nil {
		return ConnectionChannels{}, err
	}
	if r.SrcChannel == "" || r.DstChannel == "" {
		return ConnectionChannels{}, fmt.Errorf("could not load channels")
	}

	if err := r.WriteRelayerStatus(status); err != nil {
		return ConnectionChannels{}, err
	}

	return ConnectionChannels{
		Src: r.SrcChannel,
		Dst: r.DstChannel,
	}, nil
}

// waitForValidRollappHeight waits for the rollapp height to be greater than 2 otherwise
// it will fail to create clients.
func WaitForValidRollappHeight(seq *sequencer.Sequencer) error {
	spinner, _ := pterm.DefaultSpinner.Start("waiting for rollapp height to be reach 1")
	timeout := time.After(20 * time.Second)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	logger := logging.GetRollerLogger(seq.RlpCfg.Home)
	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for rollapp height to reach 1")
		case <-ticker.C:
			rollappHeightStr, err := seq.GetRollappHeight()
			if err != nil {
				logger.Printf("getting rollapp height, %s", err.Error())
				continue
			}
			rollappHeight, err := strconv.Atoi(rollappHeightStr)
			if err != nil {
				spinner.Fail("converting rollapp height to int, %s", err.Error())
				continue
			}

			if rollappHeight >= 1 {
				spinner.Success("rollapp has reached the necessary height")
				return nil
			}
			continue
		}
	}
}

func (r *Relayer) getCreateConnectionCmd(unbondingTime string) *exec.Cmd {
	args := []string{"tx", "connection", "--max-clock-drift", "70m"}
	args = append(args, r.getRelayerDefaultArgs()...)
	return exec.Command(consts.Executables.Relayer, args...)
}

type StakingParamsResponse struct {
	BondDenom         string `json:"bond_denom"`
	HistoricalEntries uint32 `json:"historical_entries"`
	MaxEntries        uint32 `json:"max_entries"`
	MaxValidators     uint32 `json:"max_validators"`
	MinCommissionRate string `json:"min_commission_rate"`
	UnbondingTime     string `json:"unbonding_time"`
}

func getHubStakingParams(hd consts.HubData) (*StakingParamsResponse, error) {
	cmd := exec.Command(
		consts.Executables.Dymension,
		"q",
		"staking",
		"params",
		"--node",
		hd.RpcUrl,
		"--chain-id",
		hd.ID,
		"--output",
		"json",
	)

	out, err := bash.ExecCommandWithStdout(cmd)
	if err != nil {
		return nil, err
	}

	var stakingParams StakingParamsResponse
	err = json.Unmarshal(out.Bytes(), &stakingParams)
	if err != nil {
		return nil, err
	}

	return &stakingParams, nil
}

// func (r *Relayer) getTxLinkCmd(override bool, unbondingTime time.Duration) *exec.Cmd {
// 	args := []string{
// 		"tx",
// 		"link",
// 		consts.DefaultRelayerPath,
// 		"--src-port",
// 		"transfer",
// 		"--dst-port",
// 		"transfer",
// 		"--version",
// 		"ics20-1",
// 		"--max-clock-drift",
// 		"70m",
// 		"--client-tp",
// 		unbondingTime.String(),
// 	}
// 	if override {
// 		args = append(args, "--override")
// 	}
// 	args = append(args, r.getRelayerDefaultArgs()...)
// 	cmd := exec.Command(consts.Executables.Relayer, args...)
// 	fmt.Println(cmd.String())
//
// 	return cmd
// }

func (r *Relayer) getCreateChannelCmd(override bool) *exec.Cmd {
	args := []string{"tx", "channel", "--timeout", "60s", "--debug"}
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
		filepath.Join(r.RollerHome, consts.ConfigDirName.Relayer),
	}
}
