package upgrades

import (
	"github.com/dymensionxyz/roller/utils/config"
	"github.com/dymensionxyz/roller/utils/roller"
	"github.com/dymensionxyz/roller/utils/sequencer"
)

// TODO: most of upgrades are associated to dymint, dymint upgrades could be extracted into separate struct
// based on dymint version and added to the upgrade modules to reduce code duplication
var WasmRollappUpgradeModules = []Version{
	{
		VersionIdentifier: "v1.0.0-rc05",
		Modules: []UpgradeModule{
			{
				Name:           "dymint",
				ConfigFilePath: sequencer.GetDymintFilePath(roller.GetRootDir()),
				Values: VersionValues{
					NewValues: []config.PathValue{
						{
							Path:  "p2p_persistent_nodes",
							Value: "",
						},
						{
							Path:  "p2p_blocksync_enabled",
							Value: "true",
						},
						{
							Path:  "p2p_blocksync_block_request_interval",
							Value: "30s",
						},
						{
							Path:  "batch_acceptance_attempts",
							Value: "5",
						},
					},
					UpgradeableValues: []UpgradeableValue{
						{
							OldValuePath: "p2p_gossiped_blocks_cache_size",
							NewValuePath: "p2p_gossip_cache_size",
							Value:        50,
						},
						{
							OldValuePath: "p2p_bootstrap_time",
							NewValuePath: "p2p_bootstrap_retry_time",
						},
						{
							OldValuePath: "p2p_advertising",
							NewValuePath: "p2p_advertising_enabled",
						},
					},
					DeprecatedValues: []string{
						"aggregator",
					},
				},
			},
		},
	},
}
