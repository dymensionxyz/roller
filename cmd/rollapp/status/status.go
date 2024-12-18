package status

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/utils/dymint"
	"github.com/dymensionxyz/roller/utils/healthagent"
	"github.com/dymensionxyz/roller/utils/roller"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show the status of the sequencer on the local machine.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String()
			rollerConfig, err := roller.LoadConfig(home)
			if err != nil {
				fmt.Println("failed to load config:", err)
				return
			}

			nodeID, err := dymint.GetNodeID(home)
			if err != nil {
				fmt.Println("failed to retrieve dymint node id:", err)
				return
			}

			ok, msg := healthagent.IsEndpointHealthy("http://localhost:26657/health")
			if !ok {
				// TODO: use options pattern, this is ugly af
				PrintOutput(rollerConfig, true, false, true, false, nodeID)
				fmt.Println("Unhealthy Message: ", msg)
				return
			}

			PrintOutput(rollerConfig, true, true, true, true, nodeID)
		},
	}
	return cmd
}

func PrintOutput(
	rlpCfg roller.RollappConfig,
	withBalance,
	withEndpoints,
	withProcessInfo,
	isHealthy bool,
	dymintNodeID string,
) {
	rollappDirPath := filepath.Join(rlpCfg.Home, consts.ConfigDirName.Rollapp)

	seq := sequencer.GetInstance(rlpCfg)

	var msg string
	if isHealthy {
		msg = pterm.DefaultBasicText.WithStyle(
			pterm.
				FgGreen.ToStyle(),
		).Sprintf("💈 The Rollapp %s is running on your local machine!", rlpCfg.NodeType)
	} else {
		msg = pterm.DefaultBasicText.WithStyle(
			pterm.
				FgRed.ToStyle(),
		).Sprintf(
			"❗ The Rollapp %s is in unhealthy state. Please check the logs for more information.",
			rlpCfg.NodeType,
		)
	}

	fmt.Println(msg)
	pterm.Println()
	fmt.Printf(
		"💈 RollApp ID: %s\n", pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
			Sprint(rlpCfg.RollappID),
	)
	fmt.Printf(
		"💈 Keyring Backend: %s\n", pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
			Sprint(rlpCfg.KeyringBackend),
	)

	fmt.Printf(
		"💈 Node ID: %s\n", pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
			Sprint(dymintNodeID),
	)

	if withEndpoints {
		pterm.DefaultSection.WithIndentCharacter("💈").
			Println("Endpoints:")
		if rlpCfg.RollappVMType == "evm" {
			fmt.Printf("EVM RPC: http://0.0.0.0:%v\n", seq.JsonRPCPort)
		}
		fmt.Printf("Node RPC: http://0.0.0.0:%v\n", seq.RPCPort)
		fmt.Printf("Rest API: http://0.0.0.0:%v\n", seq.APIPort)
	}

	pterm.DefaultSection.WithIndentCharacter("💈").
		Println("Filesystem Paths:")
	fmt.Println("Rollapp root dir: ", rollappDirPath)

	if withProcessInfo {
		pterm.DefaultSection.WithIndentCharacter("💈").
			Println("Process Info:")
		fmt.Println("OS:", runtime.GOOS)
		fmt.Println("Architecture:", runtime.GOARCH)
	}

	if isHealthy {
		seqAddrData, err := sequencerutils.GetSequencerData(rlpCfg)
		daManager := datalayer.NewDAManager(consts.Celestia, rlpCfg.Home, rlpCfg.KeyringBackend)
		celAddrData, errCel := daManager.GetDAAccData(rlpCfg)
		if err != nil {
			return
		}

		if err != nil {
			return
		}
		pterm.DefaultSection.WithIndentCharacter("💈").
			Println("Wallet Info:")
		fmt.Println("Sequencer Address:", seqAddrData[0].Address)
		if withBalance && rlpCfg.NodeType == "sequencer" {
			fmt.Println("Sequencer Balance:", seqAddrData[0].Balance.String())
		}

		if errCel != nil {
			pterm.Error.Println("failed to retrieve DA address")
			return
		}

		fmt.Println("Da Address:", celAddrData[0].Address)
		if withBalance && rlpCfg.NodeType == "sequencer" && rlpCfg.HubData.ID != "mock" {
			fmt.Println("Da Balance:", celAddrData[0].Balance.String())
		}
	}
}
