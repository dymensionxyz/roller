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
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/pterm/pterm"

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

func ExecCommandWithStdout(cmd *exec.Cmd) (*bytes.Buffer, error) {
	var stderr bytes.Buffer
	var stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		return &stderr, fmt.Errorf("command execution failed: %w, stderr: %s", err, stderr.String())
	}
	return &stdout, nil
}

func ExecCommandWithStdErr(cmd *exec.Cmd) (*bytes.Buffer, error) {
	var stderr bytes.Buffer
	var stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		return &stdout, fmt.Errorf("command execution failed: %w, stderr: %s", err, stderr.String())
	}
	return &stderr, nil
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

func ExecCmdFollow(
	doneChan chan<- error,
	ctx context.Context,
	cmd *exec.Cmd,
	promptResponses map[string]string,
) error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	// handle signals
	go func() {
		for {
			select {
			case sig := <-sigChan:
				pterm.Info.Println("received signal: ", sig)
				if cmd.Process != nil {
					_ = cmd.Process.Signal(sig)
					doneChan <- fmt.Errorf("received signal: %s", sig)
				}
			case <-ctx.Done():
				_ = cmd.Process.Signal(syscall.SIGTERM)
			}
		}
	}()

	// Use a WaitGroup to wait for both stdout and stderr to be processed
	var wg sync.WaitGroup
	wg.Add(2)

	// Channel to capture any errors from stdout or stderr
	errChan := make(chan error, 2)

	go handlePrompts(stdin, promptResponses)

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

	go func() {
		<-ctx.Done()
		if cmd.Process != nil {
			pterm.Info.Println("killing process: ", cmd.Process.Pid)
			err = cmd.Process.Kill()
			if err != nil {
				pterm.Error.Println("failed to kill process: ", err)
			}
		}
	}()

	err = cmd.Wait()
	if err != nil {
		return err
	}

	wg.Wait()
	close(errChan)

	// Check for any scanning errors
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func ExecCommandWithInteractions(cmdName string, args ...string) error {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, cmdName, args...)

	// Use the current process's standard input, output, and error
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting command: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("command finished with error: %w", err)
	}

	return nil
}

// TODO: add options: withcustomprompttext
func ExecCommandWithInput(
	home string,
	cmd *exec.Cmd,
	text string,
	promptText ...string,
) (string, error) {
	var pt string
	if len(promptText) == 0 {
		pt = "do you want to continue?"
	} else {
		pt = promptText[0]
	}
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
	//nolint:errcheck
	defer stdin.Close()
	//nolint:errcheck
	defer stdout.Close()
	//nolint:errcheck
	defer stderr.Close()

	scanner := bufio.NewScanner(io.MultiReader(stdout, stderr))
	var output strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)
		// nocheck: errcheck
		output.WriteString(line + "\n")

		if strings.Contains(line, text) {
			shouldContinue, err := pterm.DefaultInteractiveConfirm.WithDefaultText(pt).
				WithDefaultValue(false).
				Show()
			if err != nil {
				return "", err
			}

			if shouldContinue {
				if _, err := stdin.Write([]byte("y\n")); err != nil {
					return "", err
				}
			} else {
				if _, err := stdin.Write([]byte("n\n")); err != nil {
					return "", err
				}
				return "", errors.New("cancelled by user")
			}

		}
	}

	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("command finished with error: %w", err)
	}

	return output.String(), nil
}

func ExtractTxHash(output string) (string, error) {
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "raw_log:") {
			fmt.Println(line)
			str := strings.Trim(line, "raw_log:")
			fmt.Println(str)
		}
	}

	for _, line := range lines {
		if strings.HasPrefix(line, "txhash:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "txhash:")), nil
		}
	}

	return "", fmt.Errorf("txhash not found in output")
}

func handlePrompts(stdin io.WriteCloser, promptResponses map[string]string) {
	// Add a small delay to ensure the command has started
	time.Sleep(100 * time.Millisecond)

	// Write all responses
	for _, response := range promptResponses {
		// nolint: errcheck, gosec
		stdin.Write([]byte(response + "\n"))
		// Small delay between responses
		time.Sleep(100 * time.Millisecond)
	}
}

func ExecuteCommandWithPrompts(
	command string,
	args []string,
	promptResponses map[string]string,
) (*bytes.Buffer, error) {
	cmd := exec.Command(command, args...)
	// Create pipes
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %v", err)
	}
	//nolint:errcheck
	defer stdin.Close()

	// Immediately write all expected responses
	go handlePrompts(stdin, promptResponses)

	// Capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("command execution failed: %v\nOutput: %s", err, string(output))
	}

	return bytes.NewBuffer(output), nil
}

// ExecuteCommandWithPromptHandler executes a command that can handle both automatic prompt responses
// and manual interventions. For prompts that require manual intervention, provide the prompt text
// in manualPrompts. For automatic responses, provide the prompt-response pairs in promptResponses.
// TODO: refactor to handle the confirmation rather than adding -y to the args
func ExecuteCommandWithPromptHandler(
	command string,
	args []string,
	promptResponses map[string]string,
	manualPrompts map[string]string,
) (*bytes.Buffer, error) {
	cmd := exec.Command(command, args...)

	// Create pipes
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %v", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("error starting command: %v", err)
	}

	// Handle automatic prompts
	go func() {
		time.Sleep(100 * time.Millisecond)
		for _, response := range promptResponses {
			//nolint:errcheck, gosec
			stdin.Write([]byte(response + "\n"))
			time.Sleep(100 * time.Millisecond)
		}
	}()

	// Capture output
	scanner := bufio.NewScanner(io.MultiReader(stdout, stderr))
	var output strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)
		output.WriteString(line + "\n")

		// Check for manual prompts
		for promptText, question := range manualPrompts {
			if strings.Contains(line, promptText) {
				shouldContinue, err := pterm.DefaultInteractiveConfirm.
					WithDefaultText(question).
					WithDefaultValue(false).
					Show()
				if err != nil {
					return nil, err
				}

				if !shouldContinue {
					return nil, errors.New("cancelled by user")
				}

				args = append(args, "-y")
				out, err := ExecuteCommandWithPrompts(command, args, promptResponses)
				if err != nil {
					return nil, err
				}
				return out, nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading output: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("command finished with error: %w", err)
	}

	return bytes.NewBuffer([]byte(output.String())), nil
}
