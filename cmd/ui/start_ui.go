package ui

import (
	"fmt"
	_ "github.com/dymensionxyz/roller/client/statik"
	"github.com/dymensionxyz/roller/management_web_service"
	webtypes "github.com/dymensionxyz/roller/management_web_service/types"
	"github.com/spf13/cobra"
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

			management_web_service.StartManagementWebService(webtypes.Config{
				IP:      ipAddress,
				Port:    port,
				ChainID: "",
			})
		},
	}

	cmd.Flags().String(flagIp, defaultIP, "IP address to bind Web service to")
	cmd.Flags().Uint16(flagPort, defaultPort, "port to bind Web service to")
	cmd.Flags().Bool(flagConfirmUnsafeExpose, false, "must supplied when IP address is not 127.0.0.1/localhost, you have change to lost all of your balance with this option, make sure you know what you are doing")

	return cmd
}
