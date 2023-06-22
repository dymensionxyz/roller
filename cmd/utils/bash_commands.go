package utils

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

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

func GetRelayerDefaultFlags(root string) []string {
	return []string{
		"--src-port", "transfer", "--dst-port", "transfer", "--version", "ics20-1", "--home", root,
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

func ExecBashCommandWithOSOutput(cmd *exec.Cmd) error {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
