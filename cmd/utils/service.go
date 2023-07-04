package utils

import (
	"context"
	"log"
	"os/exec"
	"sync"
)

type ServiceConfig struct {
	Context   context.Context
	WaitGroup *sync.WaitGroup
	Logger    *log.Logger
}

// FIXME(#154): this functions have busy loop in case some process fails to start
func RunServiceWithRestart(cmd *exec.Cmd, serviceConfig ServiceConfig, options ...CommandOption) {
	go func() {
		defer serviceConfig.WaitGroup.Done()
		for {
			newCmd := exec.CommandContext(serviceConfig.Context, cmd.Path, cmd.Args[1:]...)
			for _, option := range options {
				option(newCmd)
			}
			commandExited := make(chan error, 1)
			go func() {
				serviceConfig.Logger.Printf("starting service command %s", newCmd.String())
				commandExited <- newCmd.Run()
			}()
			select {
			case <-serviceConfig.Context.Done():
				return
			case <-commandExited:
				serviceConfig.Logger.Printf("process %s exited, restarting...", newCmd.String())
				continue
			}
		}
	}()
}
