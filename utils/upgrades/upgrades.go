package upgrades

type VersionedSoftware interface {
	Version() (string, error)
}

type UpdateableSoftware interface {
	Upgrade(targetVersion string) string
}

// Software struct should be used by all components of the system
// as of 20241006, they are RollApp, Relayer and Eibc client
type Software struct {
	Name        string
	Binary      string
	LastVersion string
}

// Commit represents a git commit of the software version
type Commit string

// UpgradeInfo struct represents a version of a binary
type UpgradeInfo struct {
	TargetVersion Commit
	UpgradeValues []UpgradeModule
}

// UpgradeModule struct represents a module the configuration of which
// need to be upgraded. You can reason about the UpgradeModule as the
// different configuration files present in the component. For example,
// in RollApp there are dymint, app and config upgrade modules
type UpgradeModule struct {
	Name           string
	ConfigFilePath string
	Values         map[string]any
}

// UpgradeValues struct stores the different types of configuration values
// that might be used during an upgrade
type UpgradeValues struct {
	NewValues         UpgradeModule
	UpgradeableValues UpgradeModule
	DeprecatedValues  []string
}

func NewSoftware(name, bin, version string) *Software {
	return &Software{
		Name:        name,
		Binary:      bin,
		LastVersion: version,
	}
}
