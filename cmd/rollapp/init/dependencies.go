package initrollapp

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/dependencies"
	"github.com/dymensionxyz/roller/utils/dependencies/types"
	"github.com/dymensionxyz/roller/utils/rollapp"
	"github.com/dymensionxyz/roller/utils/roller"
)

// this function installs the dependencies for the rollapp
// it installs
// 1. dymd
// 2. rollapp binary
// 3. rollapp dependencies (DA)
func installDependencies(home, env string, rollapp rollapp.Rollapp) error {
	// install dymd
	defaultDymdDep := dependencies.DefaultDymdDependency()
	err := dependencies.InstallBinaryFromRelease(defaultDymdDep)
	if err != nil {
		return fmt.Errorf("failed to install dymd %s: %w", defaultDymdDep.Release, err)
	}

	// install rollapp and rollapp dependencies
	err = InstallRollappBinaries(home, env == "mock", rollapp)
	if err != nil {
		return fmt.Errorf("failed to install rollapp dependencies: %w", err)
	}
	return nil
}

func InstallRollappBinaries(home string, withMockDA bool, rollapp rollapp.Rollapp) error {
	c := exec.Command("sudo", "mkdir", "-p", consts.InternalBinsDir)
	_, err := bash.ExecCommandWithStdout(c)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", consts.InternalBinsDir, err)
	}

	var buildableDeps map[string]types.Dependency
	var goreleaserDeps map[string]types.Dependency

	if withMockDA {
		buildableDeps, goreleaserDeps, err = dependencies.GetMockDependencies(rollapp.VmType)
		if err != nil {
			return err
		}
	} else {
		buildableDeps, goreleaserDeps, err = dependencies.GetRollappDependencies(rollapp)
		if err != nil {
			return err
		}
	}

	defer func() {
		dir, err := os.UserHomeDir()
		if err != nil {
			return
		}
		_ = os.Chdir(dir)
	}()
	for k, dep := range goreleaserDeps {
		err := dependencies.InstallBinaryFromRelease(dep)
		if err != nil {
			return fmt.Errorf("failed to install binary from release %s: %w", k, err)
		}

	}

	for k, dep := range buildableDeps {
		err := dependencies.InstallBinaryFromRepo(dep, k)
		if err != nil {
			return fmt.Errorf("failed to build binary %s: %w", k, err)
		}
	}

	// set the versions to the current version

	rollerConfigFilePath := roller.GetConfigPath(home)
	fieldsToUpdate := map[string]any{
		"rollapp_binary_version": buildableDeps["rollapp"].Release,
	}
	err = tomlconfig.UpdateFieldsInFile(rollerConfigFilePath, fieldsToUpdate)
	if err != nil {
		return fmt.Errorf("failed to update roller config file: %w", err)
	}

	return nil
}
