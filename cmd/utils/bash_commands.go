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
)

func GetRelayerDefaultFlags(root string) []string {
	return []string{
		"--src-port", "transfer", "--dst-port", "transfer", "--version", "ics20-1", "--home", root,
	}
}

func RunCommandEvery(command string, args []string, intervalSec int) {
	go func() {
		for {
			cmd := exec.Command(command, args...)
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			err := cmd.Run()
			if err != nil {
				// get the cmd args joined by space
				fmt.Println("Cron command "+strings.Join(cmd.Args, " ")+" exited with error: ", stderr.String())
			}
			time.Sleep(time.Duration(intervalSec) * time.Second)
		}
	}()
}

func GetCommonDymdFlags(rollappConfig RollappConfig) []string {
	return []string{
		"--node", rollappConfig.HubData.RPC_URL, "--output", "json",
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

func ExecBashCmdWithOSOutput(cmd *exec.Cmd) error {
	var stderr bytes.Buffer
	cmd.Stdout = os.Stdout
	mw := io.MultiWriter(os.Stderr, &stderr)
	cmd.Stderr = mw
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("command execution failed: %w, stderr: %s", err, stderr.String())
	}
	return nil
}
