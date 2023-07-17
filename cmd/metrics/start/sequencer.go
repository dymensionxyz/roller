package start

import (
	"github.com/dymensionxyz/roller/config"
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"log"
	"strconv"
)

var (
	sequencerStatusGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "sequencer_status",
		Help: "The status of the local sequencer service",
	})
	rollappHeightGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "rollapp_height",
		Help: "The height of the local rollapp",
	})
	rollappHubHeightGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "rollapp_hub_height",
		Help: "The height on the hub of the local rollapp",
	})
)

func updateSequencerMetrics(rlpCfg config.RollappConfig, logger *log.Logger) {
	rollappHeight, err := sequencer.GetLocalRollappHeight(rlpCfg.RollappID)
	if err != nil {
		sequencerStatusGauge.Set(0)
		logger.Println(err)
		return
	}
	if rollappHeight == "-1" {
		sequencerStatusGauge.Set(0)
		return
	}
	sequencerStatusGauge.Set(1)
	rollappHeightFloat, err := strconv.ParseFloat(rollappHeight, 64)
	rollappHeightGauge.Set(rollappHeightFloat)
	hubHeight, err := sequencer.GetLocalRollappHubHeight(rlpCfg)
	if err != nil {
		logger.Println(err)
		return
	}
	rollappHubHeightFloat, err := strconv.ParseFloat(hubHeight, 64)
	rollappHubHeightGauge.Set(rollappHubHeightFloat)
}
