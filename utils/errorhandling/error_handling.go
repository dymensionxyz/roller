package errorhandling

import (
	"os"
	"os/signal"

	"github.com/pterm/pterm"
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

func RunOnInterrupt(funcToRun func()) {
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		funcToRun()
		os.Exit(0)
	}()
}
