package ui

import (
	"cosmossdk.io/errors"
	"fmt"
	httpclient "github.com/cometbft/cometbft/rpc/client/http"
	jsonrpcclient "github.com/cometbft/cometbft/rpc/jsonrpc/client"
	cosmosclient "github.com/cosmos/cosmos-sdk/client"
	hubapp "github.com/dymensionxyz/dymension/v3/app"
	_ "github.com/dymensionxyz/roller/client/statik"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/management_web_service"
	webtypes "github.com/dymensionxyz/roller/management_web_service/types"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/filesystem"
	queryutils "github.com/dymensionxyz/roller/utils/query"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"net/http"
	"strings"
	"time"
)

const (
	flagIp                  = "ip"
	flagPort                = "port"
	flagConfirmUnsafeExpose = "confirm-unsafe-expose"
)

const (
	cmdStartUi = "start-ui"
)

const (
	defaultIP   = "127.0.0.1"
	defaultPort = 8080
)

func Cmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   cmdStartUi,
		Short: "Start Web service",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			ipAddress, err := cmd.Flags().GetString(flagIp)
			ipAddress = strings.ToLower(strings.TrimSpace(ipAddress))
			if err != nil {
				panic(err)
			}
			port, err := cmd.Flags().GetUint16(flagPort)
			if err != nil {
				panic(err)
			}

			switch ipAddress {
			case "127.0.0.1":
			case "localhost":
				// valid
				break
			default:
				fmt.Println("### UNSAFE EXPOSE ###")
				fmt.Println("You are exposing your management tool to the public network.")
				fmt.Println("Make sure you know what you are doing!!!")
				time.Sleep(5 * time.Second)
				if !cmd.Flags().Changed(flagConfirmUnsafeExpose) {
					fmt.Printf("Please confirm exposure with --%s\n", flagConfirmUnsafeExpose)
					return
				}
			}

			rollerHome, _ := filesystem.ExpandHomePath(cmd.Flag(utils.FlagNames.Home).Value.String())
			rollappConfig, err := tomlconfig.LoadRollerConfig(rollerHome)
			if err != nil {
				pterm.Error.Printf("failed to load rollapp config: %v\n", err)
				return
			}

			var hubQueryClients *queryutils.HubQueryClients
			{
				// initialize Dymension Hub query clients
				_, tendermintRpcHttpClient, err := getTendermintClient(rollappConfig.HubData.RPC_URL)
				if err != nil {
					pterm.Error.Println("failed to initialize Tendermint client for Dymension Hub:", err)
					return
				}

				hubEncodingCfg := hubapp.MakeEncodingConfig()

				hubClientCtx := cosmosclient.Context{}.
					WithCodec(hubEncodingCfg.Codec).
					WithInterfaceRegistry(hubEncodingCfg.InterfaceRegistry).
					WithTxConfig(hubEncodingCfg.TxConfig).
					WithLegacyAmino(hubEncodingCfg.Amino).
					WithClient(tendermintRpcHttpClient)
				hubQueryClients = queryutils.NewHubQueryClients(hubClientCtx)
			}

			management_web_service.StartManagementWebService(webtypes.Config{
				CobraCmd:        cmd,
				IP:              ipAddress,
				Port:            port,
				HubQueryClients: hubQueryClients,
				RollerHome:      rollerHome,
				RollappConfig:   rollappConfig,
				WhaleAccount:    "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			})
		},
	}

	cmd.Flags().String(flagIp, defaultIP, "IP address to bind Web service to")
	cmd.Flags().Uint16(flagPort, defaultPort, "port to bind Web service to")
	cmd.Flags().Bool(flagConfirmUnsafeExpose, false, "must supplied when IP address is not 127.0.0.1/localhost, you have change to lost all of your balance with this option, make sure you know what you are doing")

	return cmd
}

func getTendermintClient(rpc string) (httpClient26657 *http.Client, tendermintRpcHttpClient *httpclient.HTTP, err error) {
	httpClient26657, err = jsonrpcclient.DefaultHTTPClient(rpc)
	if err != nil {
		return
	}
	httpTransport, ok := (httpClient26657.Transport).(*http.Transport)
	if !ok {
		err = fmt.Errorf("invalid HTTP Transport: %T", httpTransport)
		return
	}
	httpTransport.MaxConnsPerHost = 20
	tendermintRpcHttpClient, err = httpclient.NewWithClient(rpc, "/websocket", httpClient26657)
	if err != nil {
		return
	}
	err = tendermintRpcHttpClient.Start()
	if err != nil {
		err = errors.Wrap(err, "failed to start tendermint rpc http client")
		return
	}

	return
}
