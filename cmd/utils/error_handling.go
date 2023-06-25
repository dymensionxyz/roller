package utils

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"bytes"
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

func ExecBashCommand(cmd *exec.Cmd) (bytes.Buffer, error) {
	var stderr bytes.Buffer
	var stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		return stdout, fmt.Errorf("command execution failed: %w, stderr: %s", err, stderr.String())
	}
	return stdout, nil
}

func ExecBashCmdWithOSOutput(cmd *exec.Cmd, options ...CommandOption) error {
	for _, option := range options {
		option(cmd)
	}
	var stderr bytes.Buffer
	outmw := io.MultiWriter(cmd.Stdout, os.Stdout)
	cmd.Stdout = outmw
	errmw := io.MultiWriter(os.Stderr, &stderr, cmd.Stderr)
	cmd.Stderr = errmw
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("command execution failed: %w, stderr: %s", err, stderr.String())
	}
	return nil
}
