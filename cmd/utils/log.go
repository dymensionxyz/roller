package utils

import (
	"io"
	"log"
	"os/exec"
	"path/filepath"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config"
)

func GetRollerLogger(home string) *log.Logger {
	return GetLogger(filepath.Join(home, "roller.log"))
}

func WithLogging(logFile string) bash.CommandOption {
	return func(cmd *exec.Cmd) {
		logger := GetLogger(logFile)
		cmd.Stdout = logger.Writer()
		cmd.Stderr = logger.Writer()
	}
}

func WithLoggerLogging(logger *log.Logger) bash.CommandOption {
	return func(cmd *exec.Cmd) {
		cmd.Stdout = logger.Writer()
		cmd.Stderr = logger.Writer()
	}
}

func WithDiscardLogging() bash.CommandOption {
	return func(cmd *exec.Cmd) {
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
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

func GetSequencerLogPath(rollappConfig config.RollappConfig) string {
	return filepath.Join(rollappConfig.Home, consts.ConfigDirName.Rollapp, "rollapp.log")
}

func GetRelayerLogPath(home string) string {
	return filepath.Join(home, consts.ConfigDirName.Relayer, "relayer.log")
}

func GetDALogFilePath(home string) string {
	return filepath.Join(home, consts.ConfigDirName.DALightNode, "light_client.log")
}
