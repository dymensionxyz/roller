package start

// import (
// 	"fmt"
// 	"math/big"
// 	"os"
// 	"os/exec"
// 	"path/filepath"
// 	"strings"
//
// 	"github.com/spf13/cobra"
//
// 	"github.com/dymensionxyz/roller/config"
// 	"github.com/dymensionxyz/roller/sequencer"
// )
//
// // TODO: Test sequencing on 35-C and update the price
// var OneDaySequencePrice = big.NewInt(1)
//
// var (
// 	RollappDirPath string
// 	LogPath        string
// )
//
// func Cmd() *cobra.Command {
// 	cmd := &cobra.Command{
// 		Use:   "start",
// 		Short: "Show the status of the sequencer on the local machine.",
// 		Run: func(cmd *cobra.Command, args []string) {
// 		},
// 	}
// 	return cmd
// }
//
// func printOutput(rlpCfg config.RollappConfig, cmd *exec.Cmd) {
// 	seq := sequencer.GetInstance(rlpCfg)
// 	fmt.Println("💈 The Rollapp sequencer is running on your local machine!")
// 	fmt.Println("💈 Endpoints:")
//
// 	fmt.Printf("💈 EVM RPC: http://0.0.0.0:%v\n", seq.JsonRPCPort)
// 	fmt.Printf("💈 Node RPC: http://0.0.0.0:%v\n", seq.RPCPort)
// 	fmt.Printf("💈 Rest API: http://0.0.0.0:%v\n", seq.APIPort)
//
// 	fmt.Println("💈 Log file path: ", LogPath)
// 	fmt.Println("💈 Rollapp root dir: ", RollappDirPath)
// 	fmt.Println("💈 PID: ", cmd.Process.Pid)
// }
//
// func createPidFile(path string, cmd *exec.Cmd) error {
// 	pidPath := filepath.Join(path, "rollapp.pid")
// 	file, err := os.Create(pidPath)
// 	if err != nil {
// 		fmt.Println("Error creating file:", err)
// 		return err
// 	}
// 	// nolint errcheck
// 	defer file.Close()
//
// 	pid := cmd.Process.Pid
// 	_, err = file.WriteString(fmt.Sprintf("%d", pid))
// 	if err != nil {
// 		fmt.Println("Error writing to file:", err)
// 		return err
// 	}
//
// 	return nil
// }
//
// func parseError(errMsg string) string {
// 	lines := strings.Split(errMsg, "\n")
// 	if len(lines) > 0 &&
// 		lines[0] == "Error: failed to initialize database: resource temporarily unavailable" {
// 		return "The Rollapp sequencer is already running on your local machine. Only one sequencer can run at any given time."
// 	}
// 	return errMsg
// }
