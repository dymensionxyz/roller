package upgrades

import (
	"github.com/dymensionxyz/roller/utils/config"
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
	Name                 string
	Binary               string
	CurrentVersion       string
	CurrentVersionCommit string
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
	Values         VersionValues
}

type UpgradeableValue struct {
	OldValuePath string
	NewValuePath string
	Value        any
}

// VersionValues struct stores the different types of configuration values
// that might be used during an upgrade
type VersionValues struct {
	NewValues         []config.PathValue
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

type RollappUpgrade struct {
	RollappType string
	Software
}
