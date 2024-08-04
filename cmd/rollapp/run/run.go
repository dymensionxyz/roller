package run

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"github.com/dymensionxyz/roller/sequencer"
	globalutils "github.com/dymensionxyz/roller/utils"
	rollapputils "github.com/dymensionxyz/roller/utils/rollapp"
	sequencerutils "github.com/dymensionxyz/roller/utils/sequencer"
)

// TODO: Test sequencing on 35-C and update the price
var OneDaySequencePrice = big.NewInt(1)

var (
	RollappDirPath string
	LogPath        string
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Initialize RollApp locally",
		Long:  ``,
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := initconfig.AddFlags(cmd)
			if err != nil {
				pterm.Error.Println("failed to add flags")
				return
			}

			home, err := globalutils.ExpandHomePath(cmd.Flag(utils.FlagNames.Home).Value.String())
			if err != nil {
				pterm.Error.Println("failed to expand home directory")
				return
			}

			rollappConfig, err := config.LoadRollerConfigFromTOML(home)
			utils.PrettifyErrorIfExists(err)

			seq := sequencer.GetInstance(rollappConfig)
			startRollappCmd := seq.GetStartCmd()

			LogPath = filepath.Join(
				rollappConfig.Home,
				consts.ConfigDirName.Rollapp,
				"rollapputils.log",
			)
			RollappDirPath = filepath.Join(rollappConfig.Home, consts.ConfigDirName.Rollapp)

			if rollappConfig.HubData.ID == "mock" {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				go utils.RunBashCmdAsync(
					ctx, startRollappCmd, func() {
						printOutput(rollappConfig, startRollappCmd)
						err := createPidFile(RollappDirPath, startRollappCmd)
						if err != nil {
							pterm.Warning.Println("failed to create pid file")
						}
					}, parseError,
					utils.WithLogging(utils.GetSequencerLogPath(rollappConfig)),
				)
				select {}
			}

			options := []string{"sequencer", "fullnode"}
			nodeType, _ := pterm.DefaultInteractiveSelect.
				WithDefaultText("select the settlement layer backend").
				WithOptions(options).
				Show()

			if nodeType == "sequencer" {
				pterm.Info.Println("getting the existing sequencer address ")

				hubSeqKC := utils.KeyConfig{
					Dir:         filepath.Join(rollappConfig.Home, consts.ConfigDirName.HubKeys),
					ID:          consts.KeysIds.HubSequencer,
					ChainBinary: consts.Executables.Dymension,
					Type:        config.SDK_ROLLAPP,
				}
				seqAddrInfo, err := utils.GetAddressInfoBinary(hubSeqKC, hubSeqKC.ChainBinary)
				if err != nil {
					return
				}
				seqAddrInfo.Address = strings.TrimSpace(seqAddrInfo.Address)
				balance, err := utils.QueryBalance(
					utils.ChainQueryConfig{
						Denom:  consts.Denoms.Hub,
						RPC:    rollappConfig.HubData.RPC_URL,
						Binary: consts.Executables.Dymension,
					}, seqAddrInfo.Address,
				)
				if err != nil {
					return
				}
				minBond, _ := sequencerutils.GetMinSequencerBond()
				isAddrFunded := balance.Amount.Cmp(minBond.Amount.BigInt()) == 1

				isPrimarySequencer, err := rollapputils.IsPrimarySequencer(
					seqAddrInfo.Address,
					rollappConfig.RollappID,
				)
				if err != nil {
					fmt.Println(err)
				}

				if !isAddrFunded {
					pterm.DefaultSection.WithIndentCharacter("🔔").
						Println("Please fund the addresses below to register and run the sequencer.")
					seqAddrInfo.Print(utils.WithName())
					proceed, _ := pterm.DefaultInteractiveConfirm.WithDefaultValue(true).
						WithDefaultText(
							"press enter when funded",
						).Show()

					if !proceed {
						return
					}
				}

				if isPrimarySequencer {
					pterm.Info.Printf(
						"the hub_sequencer address matches the initial sequencer address of the %s\n",
						rollappConfig.RollappID,
					)
					pterm.Info.Println(
						"checking whether sequencer is already registered",
						rollappConfig.RollappID,
					)

					seq, err := rollapputils.GetRegisteredSequencers(rollappConfig.RollappID)
					if err != nil {
						pterm.Error.Println("failed to retrieve registered sequencers: ", err)
					}

					isInitialSequencerRegistered := sequencerutils.IsRegisteredAsSequencer(
						seq.Sequencers,
						seqAddrInfo.Address,
					)
					if !isInitialSequencerRegistered {
						pterm.Info.Println(
							"primary sequencer is not registered for ",
							rollappConfig.RollappID,
						)

						err = sequencerutils.Register(rollappConfig)
						if err != nil {
							pterm.Error.Println("failed to register sequencer: ", err)
							return
						}
					}
					pterm.Info.Printf(
						"%s ( %s ) is registered as a sequencer for %s",
						seqAddrInfo.Name,
						seqAddrInfo.Address,
						rollappConfig.RollappID,
					)

					return
				}
			}
		},
	}

	return cmd
}

func printOutput(rlpCfg config.RollappConfig, cmd *exec.Cmd) {
	seq := sequencer.GetInstance(rlpCfg)
	pterm.DefaultSection.WithIndentCharacter("💈 ").
		Println("The Rollapp sequencer is running on your local machine!")
	fmt.Println("💈 Endpoints:")

	fmt.Printf("EVM RPC: http://127.0.0.1:%v\n", seq.JsonRPCPort)
	fmt.Printf("Node RPC: http://127.0.0.1:%v\n", seq.RPCPort)
	fmt.Printf("Rest API: http://127.0.0.1:%v\n", seq.APIPort)

	fmt.Println("💈 Log file path: ", LogPath)
	fmt.Println("💈 Rollapp root dir: ", RollappDirPath)
	fmt.Printf(
		"💈 PID: %d (saved in %s)\n",
		cmd.Process.Pid,
		filepath.Join(rlpCfg.Home, "rollapputils.pid"),
	)
}

func createPidFile(path string, cmd *exec.Cmd) error {
	pidPath := filepath.Join(path, "rollapputils.pid")
	file, err := os.Create(pidPath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return err
	}
	// nolint errcheck
	defer file.Close()

	pid := cmd.Process.Pid
	_, err = file.WriteString(fmt.Sprintf("%d", pid))
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return err
	}

	return nil
}

func parseError(errMsg string) string {
	lines := strings.Split(errMsg, "\n")
	if len(lines) > 0 &&
		lines[0] == "Error: failed to initialize database: resource temporarily unavailable" {
		return "The Rollapp sequencer is already running on your local machine. Only one sequencer can run at any given time."
	}
	return errMsg
}
