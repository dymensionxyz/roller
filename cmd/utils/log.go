package utils

import (
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"log"
	"os/exec"
	"path/filepath"
)

func GetRollerLogger(home string) *log.Logger {
	return GetLogger(filepath.Join(home, "roller.log"))
}

func WithLogging(logFile string) CommandOption {
	return func(cmd *exec.Cmd) {
		logger := GetLogger(logFile)
		cmd.Stdout = logger.Writer()
		cmd.Stderr = logger.Writer()
	}
}

func GetLogger(filepath string) *log.Logger {
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
