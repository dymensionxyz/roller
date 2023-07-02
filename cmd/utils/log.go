package utils

import (
	"github.com/dymensionxyz/roller/cmd/consts"
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

func GetSequencerLogPath(rollappConfig RollappConfig) string {
	return filepath.Join(rollappConfig.Home, consts.ConfigDirName.Rollapp, "rollapp.log")
}

func GetRelayerLogPath(config RollappConfig) string {
	return filepath.Join(config.Home, consts.ConfigDirName.Relayer, "relayer.log")
}

func GetDALogFilePath(home string) string {
	return filepath.Join(home, consts.ConfigDirName.DALightNode, "light_client.log")
}
