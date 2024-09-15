package logs

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/filesystem"
)

func RollappCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Follow the logs for rollapp and da light client",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()

			rollerData, err := tomlconfig.LoadRollerConfig(home)
			if err != nil {
				pterm.Error.Println("failed to load roller config file", err)
				return
			}

			raLogFilePath := filepath.Join(
				rollerData.Home,
				consts.ConfigDirName.Rollapp,
				"rollapp.log",
			)

			daLogFilePath := filepath.Join(
				rollerData.Home,
				consts.ConfigDirName.DALightNode,
				"light_client.log",
			)
			pterm.Info.Println("Follow the logs for rollapp: ", raLogFilePath)
			pterm.Info.Println("Follow the logs for da light client: ", daLogFilePath)

			errChan := make(chan error, 2)
			doneChan := make(chan bool)

			go func() {
				err := filesystem.TailFile(raLogFilePath)
				if err != nil {
					pterm.Error.Println("failed to tail file", err)
					errChan <- fmt.Errorf("failed to tail RA file: %w", err)
					return
				}
			}()
			go func() {
				err := filesystem.TailFile(daLogFilePath)
				if err != nil {
					pterm.Error.Println("failed to tail file", err)
					errChan <- fmt.Errorf("failed to tail DA file: %w", err)
					return
				}
			}()

			// Keep the program running
			go func() {
				time.Sleep(time.Hour) // Adjust this duration as needed
				doneChan <- true
			}()

			select {
			case err := <-errChan:
				pterm.Error.Println(err)
				os.Exit(1)
			case <-doneChan:
				pterm.Info.Println("finished")
			}
		},
	}
	return cmd
}
