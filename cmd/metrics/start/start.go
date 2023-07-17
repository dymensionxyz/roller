package start

import (
	"fmt"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

func getMetricsLogger(rollerHome string) *log.Logger {
	return utils.GetLogger(filepath.Join(rollerHome, "metrics.log"))
}

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start a prometheus metrics server for the local rollapp",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("💈 The prometheus metrics server is running successfully on http://localhost:2112/metrics")
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rollappConfig, err := config.LoadConfigFromTOML(home)
			utils.PrettifyErrorIfExists(err)
			go func() {
				logger := getMetricsLogger(rollappConfig.Home)
				for {
					updateRollappMetrics(rollappConfig, logger)
					time.Sleep(5 * time.Second)
				}
			}()

			utils.PrettifyErrorIfExists(startMetricsServer())
		},
	}
	return cmd
}

func startMetricsServer() error {
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(":2112", nil)
	return err
}

func updateRollappMetrics(rlpCfg config.RollappConfig, logger *log.Logger) {
	updateSequencerMetrics(rlpCfg, logger)
	updateDALCMetrics(rlpCfg)
	updateRelayerMetrics(rlpCfg, logger)
}
