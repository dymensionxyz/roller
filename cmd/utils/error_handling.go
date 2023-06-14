package utils

import (
	"fmt"
	"os"
	"os/exec"

	"bytes"
	"errors"
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

func RunBashCmdAsync(cmd *exec.Cmd, printOutput func(), parseError func(errMsg string) string) {
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Start()
	if err != nil {
		errMsg := parseError(stderr.String())
		PrettifyErrorIfExists(errors.New(errMsg))
	}
	printOutput()
	err = cmd.Wait()
	if err != nil {
		errMsg := parseError(stderr.String())
		PrettifyErrorIfExists(errors.New(errMsg))
	}
}

func ExecBashCommand(cmd *exec.Cmd) error {
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Todo: Remove this
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("command execution failed: %w, stderr: %s", err, stderr.String())
	}

	return nil
}
