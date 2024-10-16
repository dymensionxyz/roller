package start

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/relayer"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/logging"
	"github.com/dymensionxyz/roller/utils/rollapp"
	"github.com/dymensionxyz/roller/utils/roller"
	sequencerutils "github.com/dymensionxyz/roller/utils/sequencer"
)

// TODO: Test relaying on 35-C and update the prices

const (
	flagOverride = "override"
)

type Config struct {
	Paths struct {
		HubRollapp struct {
			Dst struct {
				ChainID string `yaml:"chain-id"`
			} `yaml:"dst"`
			Src struct {
				ChainID string `yaml:"chain-id"`
			} `yaml:"src"`
		} `yaml:"hub-rollapp"`
	} `yaml:"paths"`
}

func Cmd() *cobra.Command {
	relayerStartCmd := &cobra.Command{
		Use:   "start",
		Short: "Start the relayer process interactively.",
		Long: `Start the relayer process interactively.

Consider using 'services' if you want to run a 'systemd' service instead.
`,
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String()
			rlyConfigPath := filepath.Join(
				home,
				consts.ConfigDirName.Relayer,
				"config",
				"config.yaml",
			)

			data, err := os.ReadFile(rlyConfigPath)
			if err != nil {
				fmt.Printf("Error reading YAML file: %v\n", err)
				return
			}

			var rlyConfig Config
			err = yaml.Unmarshal(data, &rlyConfig)
			if err != nil {
				fmt.Printf("Error unmarshaling YAML: %v\n", err)
				return
			}

			// src is Hub, dst is RollApp
			raChainID := rlyConfig.Paths.HubRollapp.Dst.ChainID
			hubChainID := rlyConfig.Paths.HubRollapp.Src.ChainID

			_, hd, found := roller.FindHubDataByID(consts.Hubs, hubChainID)
			if !found {
				pterm.Error.Println("Hub Data not found for ", hubChainID)
				runCustomHub, _ := pterm.DefaultInteractiveConfirm.WithDefaultText("would you like to provide custom hub?").
					WithDefaultValue(false).
					Show()

				if !runCustomHub {
					pterm.Error.Println("cancelled by user")
					return
				}

				id, _ := pterm.DefaultInteractiveTextInput.WithDefaultText("provide hub chain id").
					Show()
				rpcUrl, _ := pterm.DefaultInteractiveTextInput.WithDefaultText(
					"provide hub rpc endpoint (including port, example: http://dym.dev:26657)",
				).Show()
				restUrl, _ := pterm.DefaultInteractiveTextInput.WithDefaultText(
					"provide hub rest api endpoint (including port, example: http://dym.dev:1318)",
				).Show()
				gasPrice, _ := pterm.DefaultInteractiveTextInput.WithDefaultText("provide gas price").
					WithDefaultValue("2000000000").
					Show()

				id = strings.TrimSpace(id)
				rpcUrl = strings.TrimSpace(rpcUrl)
				restUrl = strings.TrimSpace(restUrl)
				gasPrice = strings.TrimSpace(gasPrice)

				hd = consts.HubData{
					API_URL:         restUrl,
					ID:              id,
					RPC_URL:         rpcUrl,
					ARCHIVE_RPC_URL: rpcUrl,
					GAS_PRICE:       gasPrice,
				}
			}

			getRaCmd := rollapp.GetRollappCmd(raChainID, hd)
			var raResponse rollapp.ShowRollappResponse

			out, err := bash.ExecCommandWithStdout(getRaCmd)
			if err != nil {
				pterm.Error.Println("failed to get rollapp: ", err)
				return
			}
			err = json.Unmarshal(out.Bytes(), &raResponse)
			if err != nil {
				pterm.Error.Println("failed to unmarshal", err)
				return
			}

			// errorhandling.RequireMigrateIfNeeded(rollappConfig)
			raRpc, err := sequencerutils.GetRpcEndpointFromChain(
				raChainID,
				hd,
			)
			if err != nil {
				return
			}
			raData := consts.RollappData{
				ID:     raChainID,
				RpcUrl: fmt.Sprintf("%s:%d", raRpc, 443),
				Denom:  raResponse.Rollapp.GenesisInfo.NativeDenom.Base,
			}

			err = VerifyRelayerBalances(raData, hd)
			if err != nil {
				pterm.Error.Println("failed to check balances", err)
				return
			}
			relayerLogFilePath := logging.GetRelayerLogPath(home)
			logger := logging.GetLogger(relayerLogFilePath)
			logFileOption := logging.WithLoggerLogging(logger)
			rly := relayer.NewRelayer(
				home,
				raChainID,
				hubChainID,
			)
			rly.SetLogger(logger)

			// TODO: relayer is initialized with both chains at this point and it should be possible
			// to construct the hub data from relayer config
			_, _, err = rly.LoadActiveChannel(raData, hd)
			errorhandling.PrettifyErrorIfExists(err)

			// override := cmd.Flag(flagOverride).Changed
			//
			// if override {
			// 	fmt.Println("💈 Overriding the existing relayer channel")
			// }

			if rly.ChannelReady() {
				fmt.Println("💈 IBC transfer channel is already established!")
				status := fmt.Sprintf(
					"Active\nrollapp: %s\n<->\nhub: %s\n",
					rly.DstChannel, rly.SrcChannel,
				)
				err := rly.WriteRelayerStatus(status)
				errorhandling.PrettifyErrorIfExists(err)
			} else {
				pterm.Error.Println("💈 No channels found, ensure you've setup the relayer")
				return
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			go bash.RunCmdAsync(
				ctx,
				rly.GetStartCmd(),
				func() {},
				func(errMessage string) string { return errMessage },
				logFileOption,
			)

			fmt.Printf(
				"💈 The relayer is running successfully on you local machine!\nChannels:\nrollapp: %s\n<->\nhub: %s\n",
				rly.SrcChannel,
				rly.DstChannel,
			)
			fmt.Println("💈 Log file path: ", relayerLogFilePath)

			select {}
		},
	}

	relayerStartCmd.Flags().
		BoolP(flagOverride, "", false, "override the existing relayer clients and channels")
	return relayerStartCmd
}

func VerifyRelayerBalances(raData consts.RollappData, hd consts.HubData) error {
	insufficientBalances, err := relayer.GetRelayerInsufficientBalances(raData, hd)
	if err != nil {
		return err
	}

	if len(insufficientBalances) != 0 {
		err = keys.PrintInsufficientBalancesIfAny(insufficientBalances)
		if err != nil {
			return err
		}
	}

	return nil
}
