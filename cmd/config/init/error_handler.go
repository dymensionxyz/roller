package initconfig

import (
	"os"

	"github.com/fatih/color"
)

func OutputCleanError(err error) {
	if err != nil {
		defer func() {
			if r := recover(); r != nil {
				os.Exit(1)
			}
		}()
		color.Red(err.Error())
		panic(err)
	}
}
