package upgrades

import (
	"github.com/dymensionxyz/roller/utils/roller"
	"github.com/dymensionxyz/roller/utils/sequencer"
)

type VersionedSoftware interface {
	Version() (string, error)
}

type UpdateableSoftware interface {
	Upgrade(targetVersion string) string
}

// Software struct should be used by all components of the system
// as of 20241006, they are RollApp, Relayer and Eibc client
type Software struct {
	Name           string
	Binary         string
	CurrentVersion string
}

// Commit represents a git commit of the software version
type Commit string

// UpgradeInfo struct represents a version of a binary
type UpgradeInfo struct {
	TargetVersion Commit
	UpgradeValues []UpgradeModule
}

// Version struct encompass all modules that are related to a specific
// component
type Version struct {
	VersionIdentifier string // version can be commit, tag or release string
	Modules           []UpgradeModule
}

// UpgradeModule struct represents a module the configuration of which
// need to be upgraded. You can reason about the UpgradeModule as the
// different configuration files present in the component. For example,
// in RollApp there are dymint, app and config upgrade modules
type UpgradeModule struct {
	Name           string
	ConfigFilePath string
	Values         VerstionValues
}

type UpgradeableValue struct {
	OldValuePath string
	NewValuePath string
	Value        any
}

// VerstionValues struct stores the different types of configuration values
// that might be used during an upgrade
type VerstionValues struct {
	NewValues         map[string]any
	UpgradeableValues []UpgradeableValue
	DeprecatedValues  []string
}

func NewSoftware(name, bin, version string) *Software {
	return &Software{
		Name:           name,
		Binary:         bin,
		CurrentVersion: version,
	}
}

type (
	RollappUpgrade Software
)

var EvmRollappUpgradeModules = []Version{
	{
		VersionIdentifier: "v2.2.1-rc05",
		Modules: []UpgradeModule{
			{
				Name:           "dymint",
				ConfigFilePath: sequencer.GetDymintFilePath(roller.GetRootDir()),
				Values: VerstionValues{
					NewValues: map[string]any{
						"p2p_persistent_nodes":                 "",
						"p2p_blocksync_enabled":                true,
						"p2p_blocksync_block_request_interval": "30s",
						"batch_acceptance_attempts":            5,
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
