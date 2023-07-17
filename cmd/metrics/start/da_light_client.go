package start

import (
	"github.com/dymensionxyz/roller/config"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var daLCStatus = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "da_lc_status",
	Help: "The status of the DA light client",
})

func updateDALCMetrics(rlpCfg config.RollappConfig) {
	damanager := datalayer.NewDAManager(rlpCfg.DA, rlpCfg.Home)
	daStatus := damanager.GetStatus(rlpCfg)
	if daStatus == "Active" {
		daLCStatus.Set(1)
	} else {
		daLCStatus.Set(0)
	}
}
