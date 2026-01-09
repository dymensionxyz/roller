package start

import (
	"context"
	"errors"
	"os/exec"
	"testing"
	"time"

	"github.com/dymensionxyz/roller/utils/bash"
)

func TestRunRollappCommandSuccess(t *testing.T) {
	t.Cleanup(func() { execCmdFollowFunc = bashExecCmdFollowWrapper })
	execCmdFollowFunc = func(done chan<- error, ctx context.Context, cmd *exec.Cmd, _ map[string]string) error {
		done <- nil
		return nil
	}

    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()

    if err := runRollappCommand(ctx, exec.Command("echo"), nil); err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
}

func TestRunRollappCommandReturnsProcessError(t *testing.T) {
	t.Cleanup(func() { execCmdFollowFunc = bashExecCmdFollowWrapper })
	execCmdFollowFunc = func(done chan<- error, ctx context.Context, cmd *exec.Cmd, _ map[string]string) error {
		return errors.New("boom")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := runRollappCommand(ctx, exec.Command("echo"), nil)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestRunRollappCommandSignals(t *testing.T) {
	t.Cleanup(func() { execCmdFollowFunc = bashExecCmdFollowWrapper })
	execCmdFollowFunc = func(done chan<- error, ctx context.Context, cmd *exec.Cmd, _ map[string]string) error {
		done <- errors.New("SIGTERM")
        return nil
    }

    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()

    if err := runRollappCommand(ctx, exec.Command("echo"), nil); err == nil {
        t.Fatalf("expected signal error")
    }
}

func TestRunRollappCommandContextCancelled(t *testing.T) {
    t.Cleanup(func() { execCmdFollowFunc = bashExecCmdFollowWrapper })
    execCmdFollowFunc = func(done chan<- error, ctx context.Context, cmd *exec.Cmd, _ map[string]string) error {
        select {
        case <-ctx.Done():
            return ctx.Err()
        }
    }

    ctx, cancel := context.WithCancel(context.Background())
    cancel()

    if err := runRollappCommand(ctx, exec.Command("echo"), nil); err == nil {
        t.Fatalf("expected context error")
    }
}

func bashExecCmdFollowWrapper(done chan<- error, ctx context.Context, cmd *exec.Cmd, prompts map[string]string) error {
    return bash.ExecCmdFollow(done, ctx, cmd, prompts)
}
