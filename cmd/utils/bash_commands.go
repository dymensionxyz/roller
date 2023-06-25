package utils

import (
	"bytes"
	"errors"
	"fmt"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"log"
	"os/exec"
	"strings"
	"time"
)

func GetRelayerDefaultFlags(root string) []string {
	return []string{
		"--src-port", "transfer", "--dst-port", "transfer", "--version", "ics20-1", "--home", root,
	}
}

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

type CommandOption func(cmd *exec.Cmd)

func RunBashCmdAsync(cmd *exec.Cmd, printOutput func(), parseError func(errMsg string) string,
	options ...CommandOption) {
	for _, option := range options {
		option(cmd)
	}
	var stderr bytes.Buffer
	mw := io.MultiWriter(cmd.Stderr, &stderr)
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

func WithLogging(logFile string) CommandOption {
	return func(cmd *exec.Cmd) {
		logger := getLogger(logFile)
		cmd.Stdout = logger.Writer()
		cmd.Stderr = logger.Writer()
	}
}

func getLogger(filepath string) *log.Logger {
	lumberjackLogger := &lumberjack.Logger{
		Filename:   filepath,
		MaxSize:    500,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}
	multiWriter := io.MultiWriter(lumberjackLogger)
	logger := log.New(multiWriter, "", log.LstdFlags)
	return logger
}
