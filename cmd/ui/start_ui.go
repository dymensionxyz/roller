package ui

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"cosmossdk.io/errors"

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
	cmd := &cobra.Command{
		Use:   cmdStartUi,
		Short: "Start Web service",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			initSimulation()
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

			{
				// ensure no eIBC-client running
				first := true
				for {
					if !first {
						time.Sleep(500 * time.Second)
					}
					first = false
					anyEIbcClient, err := management_web_service.AnyEIbcClient()
					if err != nil {
						pterm.Error.Println("failed to check eIBC-client:", err)
						continue
					}
					if anyEIbcClient {
						pterm.Error.Println("eIBC-client is running, please stop it first")
						os.Exit(1)
					}

					break
				}
			}

			// Ensure any background process should be killed upon exit.
			trapSignal(func() {
				management_web_service.Cleanup()
			})

			management_web_service.StartManagementWebService(&webtypes.Config{
				CobraCmd:        cmd,
				IP:              ipAddress,
				Port:            port,
				HubQueryClients: hubQueryClients,
				RollerHome:      rollerHome,
				RollappConfig:   rollappConfig,
				WhaleAccount:    "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue", // pseudo address
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

// trapSignal traps SIGINT and SIGTERM and calls os.Exit once a signal is received.
func trapSignal(cleanupFunc func()) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs

		if cleanupFunc != nil {
			cleanupFunc()
		}
		exitCode := 128

		switch sig {
		case syscall.SIGINT:
			exitCode += int(syscall.SIGINT)
		case syscall.SIGTERM:
			exitCode += int(syscall.SIGTERM)
		}

		os.Exit(exitCode)
	}()
}

func initSimulation() {
	sim := os.Getenv("SIMULATION")
	if sim == "" {
		return
	}
	spl := strings.Split(sim, " ")
	management_web_service.EIbcClientBinaryName = spl[0] // override
	management_web_service.SimulationStartCommand = spl[1]
	management_web_service.UseSimulation = true
}
