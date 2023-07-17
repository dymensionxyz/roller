package start

import (
	"github.com/dymensionxyz/roller/config"
	"github.com/dymensionxyz/roller/relayer"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"log"
)

var relayerStatus = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "relayer_status",
	Help: "The status of the local rollapp relayer service",
})

func updateRelayerMetrics(rlpCfg config.RollappConfig, logger *log.Logger) {
	channels, err := relayer.GetChannels(rlpCfg)
	if err != nil || channels.Src == "" {
		relayerStatus.Set(0)
		if err != nil {
			logger.Println(err)
		}
	} else {
		relayerStatus.Set(1)
	}
}
