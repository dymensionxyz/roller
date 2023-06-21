package utils

import (

	"github.com/fatih/color"
	"os"
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
