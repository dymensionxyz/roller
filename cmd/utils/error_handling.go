package utils

import (
	"os"

	"github.com/fatih/color"
)

func PrettifyErrorIfExists(err error) {
	if err != nil {
		defer func() {
			if r := recover(); r != nil {
				os.Exit(1)
			}
		}()
		color.New(color.FgRed, color.Bold).Printf("💈 %s\n", err.Error())
		panic(err)
	}
}
