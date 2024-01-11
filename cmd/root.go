package cmd

import (
	"encoding/json"
	"github.com/dymensionxyz/roller/cmd/migrate"
	"github.com/dymensionxyz/roller/cmd/run"
	"github.com/dymensionxyz/roller/cmd/services"
	"github.com/dymensionxyz/roller/cmd/tx"
	"github.com/dymensionxyz/roller/cmd/utils"
	"os"
	"os/exec"
	"strings"

	"github.com/dymensionxyz/roller/cmd/config"
	da_light_client "github.com/dymensionxyz/roller/cmd/da-light-client"
	"github.com/dymensionxyz/roller/cmd/keys"
	"github.com/dymensionxyz/roller/cmd/relayer"
	"github.com/dymensionxyz/roller/cmd/sequencer"
	"github.com/dymensionxyz/roller/cmd/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "roller",
	Short: "A simple CLI tool to spin up a RollApp",
	Long: `
Roller CLI is a tool for registering and running autonomous RollApps built with Dymension RDK. Roller provides everything you need to scaffold, configure, register, and run your RollApp.
	`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(config.Cmd())
	rootCmd.AddCommand(version.Cmd())
	rootCmd.AddCommand(da_light_client.DALightClientCmd())
	rootCmd.AddCommand(sequencer.SequencerCmd())
	rootCmd.AddCommand(relayer.Cmd())
	rootCmd.AddCommand(keys.Cmd())
	rootCmd.AddCommand(run.Cmd())
	rootCmd.AddCommand(services.Cmd())
	rootCmd.AddCommand(migrate.Cmd())
	rootCmd.AddCommand(tx.Cmd())
	rootCmd.AddCommand(test())
	utils.AddGlobalFlags(rootCmd)
}

func test() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Runs the rollapp on the local machine.",
		RunE: func(cmd *cobra.Command, args []string) error {
			resp := checkBalance()
			println(resp)
			return nil
		},
	}
	return cmd
}

type CelestiaResponse struct {
	Result struct {
		Denom  string `json:"denom"`
		Amount string `json:"amount"`
	} `json:"result"`
}

func checkBalance() string {
	cmd := "/usr/local/bin/roller_bins/celestia"
	args := []string{"state", "balance", "--node.store", "~/.roller/da-light-node"}
	output, err := exec.Command(cmd, args...).Output()

	if err != nil {
		return "Stopped, Restarting"
	}

	var resp CelestiaResponse
	err = json.Unmarshal(output, &resp)
	if err != nil {
		return "Stopped, Restarting"
	}

	if strings.TrimSpace(resp.Result.Amount) != "0" {
		return "active"
	}

	return "Stopped, Restarting"
}
