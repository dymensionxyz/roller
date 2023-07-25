package utils

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/dymensionxyz/roller/config"
)

func GetRelayerDefaultFlags(root string) []string {
	return []string{
		"--src-port", "transfer", "--dst-port", "transfer", "--version", "ics20-1", "--home", root,
	}
}

// TODO: should accept a context and cancel the command if the context is cancelled
func RunCommandEvery(command string, args []string, intervalSec int, options ...CommandOption) {
	go func() {
		for {
			cmd := exec.Command(command, args...)
			for _, option := range options {
				option(cmd)
			}
			var stderr bytes.Buffer
			errmw := io.MultiWriter(&stderr, cmd.Stderr)
			cmd.Stderr = errmw
			err := cmd.Run()
			if err != nil {
				fmt.Println("periodic command "+strings.Join(cmd.Args, " ")+" exited with error: ", stderr.String())
			}
			time.Sleep(time.Duration(intervalSec) * time.Second)
		}
	}()
}

func GetCommonDymdFlags(rollappConfig config.RollappConfig) []string {
	return []string{
		"--node", rollappConfig.HubData.RPC_URL, "--output", "json",
	}
}

type CommandOption func(cmd *exec.Cmd)

func RunBashCmdAsync(cmd *exec.Cmd, printOutput func(), parseError func(errMsg string) string,
	options ...CommandOption) {
	for _, option := range options {
		option(cmd)
	}
	if parseError == nil {
		parseError = func(errMsg string) string {
			return errMsg
		}
	}

	var stderr bytes.Buffer
	mw := io.MultiWriter(&stderr)
	if cmd.Stderr != nil {
		mw = io.MultiWriter(&stderr, cmd.Stderr)
	}
	cmd.Stderr = mw
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

func ExecBashCommandWithStdErr(cmd *exec.Cmd) (bytes.Buffer, error) {
	var stderr bytes.Buffer
	var stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		return stdout, fmt.Errorf("command execution failed: %w, stderr: %s", err, stderr.String())
	}
	return stderr, nil
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
