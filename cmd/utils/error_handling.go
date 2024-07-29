package utils

import (
	"os"
	"os/signal"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/config"
	"github.com/dymensionxyz/roller/version"
)

func PrettifyErrorIfExists(err error, printAdditionalInfo ...func()) {
	if err != nil {
		defer func() {
			if r := recover(); r != nil {
				os.Exit(1)
			}
		}()
		pterm.Error.Printf("💈 %s\n", err.Error())

		for _, printInfo := range printAdditionalInfo {
			printInfo()
		}

		panic(err)
	}
}

func RequireMigrateIfNeeded(rlpCfg config.RollappConfig) {
	currentRollerVersion := version.TrimVersionStr(version.BuildVersion)
	configRollerVersion := version.TrimVersionStr(rlpCfg.RollerVersion)
	if configRollerVersion != currentRollerVersion {
		//nolint:errcheck,gosec
		pterm.Warning.Printf(
			"💈 Your rollapp config version ('%s') is older than your"+
				" installed roller version ('%s'),"+
				" please run 'roller migrate' to update your config.\n", configRollerVersion, currentRollerVersion,
		)
		os.Exit(1)
	}
}

func RunOnInterrupt(funcToRun func()) {
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		funcToRun()
		os.Exit(0)
	}()
}
