package start

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
)

// Creates an IBC channel between the hub and the client, and return the source channel ID.
func createIBCChannelIfNeeded(rollappConfig utils.RollappConfig, logFileOption utils.CommandOption) (string, error) {
	createClientsCmd := getCreateClientsCmd(rollappConfig, rollappConfig.RollappID, rollappConfig.HubData.ID)
	fmt.Println("Creating clients...")
	if err := utils.ExecBashCmdWithOSOutput(createClientsCmd, logFileOption); err != nil {
		return "", err
	}
	dstConnectionId, err := GetDstConnectionIDFromYAMLFile(filepath.Join(rollappConfig.Home, consts.ConfigDirName.Relayer,
		"config", "config.yaml"))
	if err != nil {
		return "", err
	}
	if dstConnectionId == "" {
		createConnectionCmd := getCreateConnectionCmd(rollappConfig)
		fmt.Println("Creating connection...")
		if err := utils.ExecBashCmdWithOSOutput(createConnectionCmd, logFileOption); err != nil {
			return "", err
		}
	}
	srcChannelId, err := GetSourceChannelForConnection(dstConnectionId, rollappConfig)
	if err != nil {
		return "", err
	}
	if srcChannelId == "" {
		createChannelCmd := getCreateChannelCmd(rollappConfig)
		fmt.Println("Creating channel...")
		if err := utils.ExecBashCmdWithOSOutput(createChannelCmd, logFileOption); err != nil {
			return "", err
		}
		srcChannelId, err = GetSourceChannelForConnection(dstConnectionId, rollappConfig)
		if err != nil {
			return "", err
		}
	}
	return srcChannelId, nil
}

func getCreateChannelCmd(config utils.RollappConfig) *exec.Cmd {
	defaultRlyArgs := getRelayerDefaultArgs(config)
	args := []string{"tx", "channel", "-t", "300s", "--override"}
	args = append(args, defaultRlyArgs...)
	return exec.Command(consts.Executables.Relayer, args...)
}

func getCreateClientsCmd(rollappConfig utils.RollappConfig, srcId string, dstId string) *exec.Cmd {
	defaultRlyArgs := getRelayerDefaultArgs(rollappConfig)
	args := []string{"tx", "clients"}
	args = append(args, defaultRlyArgs...)
	return exec.Command(consts.Executables.Relayer, args...)
}

func getRelayerDefaultArgs(config utils.RollappConfig) []string {
	return []string{consts.DefaultRelayerPath, "--home", filepath.Join(config.Home, consts.ConfigDirName.Relayer)}
}

func getCreateConnectionCmd(config utils.RollappConfig) *exec.Cmd {
	defaultRlyArgs := getRelayerDefaultArgs(config)
	args := []string{"tx", "connection", "-t", "300s"}
	args = append(args, defaultRlyArgs...)
	return exec.Command(consts.Executables.Relayer, args...)
}
