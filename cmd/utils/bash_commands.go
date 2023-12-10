package utils

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"time"

	"github.com/dymensionxyz/roller/config"
)

func RunCommandEvery(ctx context.Context, command string, args []string, intervalSec int, options ...CommandOption) {
	go func() {
		for {
			cmd := exec.CommandContext(ctx, command, args...)
			for _, option := range options {
				option(cmd)
			}
			err := cmd.Run()

			if err != nil {
				cmd.Stderr.Write([]byte(fmt.Sprintf("Failed to execute command: %s, err: %s\n", cmd.String(), err)))
			}

			if ctx.Err() != nil {
				return
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

func RunBashCmdAsync(ctx context.Context, cmd *exec.Cmd, printOutput func(), parseError func(errMsg string) string,
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
		if errMsg == "" {
			PrettifyErrorIfExists(err)
		}
		PrettifyErrorIfExists(errors.New(errMsg))
	}
	printOutput()

	go func() {
		<-ctx.Done()
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

	err = cmd.Wait()
	if err != nil {
		errMsg := parseError(stderr.String())
		PrettifyErrorIfExists(errors.New(errMsg))
	}
}

func ExecBashCommandWithStdout(cmd *exec.Cmd) (bytes.Buffer, error) {
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

func ExecBashCmd(cmd *exec.Cmd, options ...CommandOption) error {
	for _, option := range options {
		option(cmd)
	}
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}
	return nil
}
