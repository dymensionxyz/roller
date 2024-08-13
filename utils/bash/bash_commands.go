package bash

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	yaml "gopkg.in/yaml.v2"

	"github.com/dymensionxyz/roller/utils/errorhandling"
)

func RunCommandEvery(
	ctx context.Context,
	command string,
	args []string,
	intervalSec int,
	options ...CommandOption,
) {
	go func() {
		for {
			cmd := exec.CommandContext(ctx, command, args...)
			for _, option := range options {
				option(cmd)
			}
			err := cmd.Run()
			if err != nil {
				_, err := cmd.Stderr.Write(
					[]byte(
						fmt.Sprintf("Failed to execute command: %s, err: %s\n", cmd.String(), err),
					),
				)
				if err != nil {
					return
				}
			}

			if ctx.Err() != nil {
				return
			}

			time.Sleep(time.Duration(intervalSec) * time.Second)
		}
	}()
}

type CommandOption func(cmd *exec.Cmd)

func RunCmdAsync(
	ctx context.Context,
	cmd *exec.Cmd,
	printOutput func(),
	parseError func(errMsg string) string,
	options ...CommandOption,
) {
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
			errorhandling.PrettifyErrorIfExists(err)
		}
		errorhandling.PrettifyErrorIfExists(errors.New(errMsg))
	}
	printOutput()

	go func() {
		<-ctx.Done()
		if cmd.Process != nil {
			err := cmd.Process.Kill()
			if err != nil {
				return
			}
		}
	}()

	err = cmd.Wait()
	if err != nil {
		errMsg := parseError(stderr.String())
		errorhandling.PrettifyErrorIfExists(errors.New(errMsg))
	}
}

func ExecCommandWithStdout(cmd *exec.Cmd) (bytes.Buffer, error) {
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

func ExecCommandWithStdErr(cmd *exec.Cmd) (bytes.Buffer, error) {
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

func ExecCmd(cmd *exec.Cmd, options ...CommandOption) error {
	for _, option := range options {
		option(cmd)
	}
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("command execution failed: %v", err)
	}
	return nil
}

func ExecCmdFollow(cmd *exec.Cmd) error {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	// Use a WaitGroup to wait for both stdout and stderr to be processed
	var wg sync.WaitGroup
	wg.Add(2)

	// Channel to capture any errors from stdout or stderr
	errChan := make(chan error, 2)

	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			errChan <- err
		}
	}()

	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			errChan <- err
		}
	}()

	// Wait for both stdout and stderr goroutines to finish
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		return err
	}

	// Check if there were any errors in the goroutines
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

// TODO: generalize

func ExecCommandWithInput(cmd *exec.Cmd) (string, error) {
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("error creating stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("error creating stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("error creating stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("error starting command: %w", err)
	}

	scanner := bufio.NewScanner(io.MultiReader(stdout, stderr))
	var txHash string
	var yamlOutput strings.Builder
	var captureYAML bool

	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)

		if strings.Contains(line, "Do you want to continue?") {
			fmt.Print("Do you want to continue? (y/n): ")
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)
			fmt.Println("input:", input)

			if input == "y" || input == "Y" {
				if _, err := stdin.Write([]byte("y\n")); err != nil {
					return "", err
				}
				captureYAML = true
			} else {
				if _, err := stdin.Write([]byte("n\n")); err != nil {
					return "", err
				}
				break
			}
		} else if captureYAML {
			yamlOutput.WriteString(line + "\n")
		}
	}

	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("command finished with error: %w", err)
	}

	// Parse YAML output
	var result map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlOutput.String()), &result); err != nil {
		return "", fmt.Errorf("error parsing YAML output: %w", err)
	}

	// Extract txhash
	if hash, ok := result["txhash"].(string); ok {
		txHash = hash
	}

	return txHash, nil
}
