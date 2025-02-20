package bash

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/pterm/pterm"
)

// ExecCmdFollowWithHandler executes a command and allows processing output through a handler function.
// The outputHandler can return true to indicate that the command should be terminated early.
func ExecCmdFollowWithHandler(
	doneChan chan<- error,
	ctx context.Context,
	cmd *exec.Cmd,
	outputHandler func(line string) bool,
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

	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println(line)
			if outputHandler != nil {
				if shouldStop := outputHandler(line); shouldStop {
					// Signal to main routine that we want to stop
					doneChan <- nil
					return
				}
			}
		}
		if err := scanner.Err(); err != nil {
			errChan <- err
		}
	}()

	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println(line)
			if outputHandler != nil {
				if shouldStop := outputHandler(line); shouldStop {
					// Signal to main routine that we want to stop
					doneChan <- nil
					return
				}
			}
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
		// Check if this was due to us killing the process
		if strings.Contains(err.Error(), "signal: killed") {
			return nil
		}
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
