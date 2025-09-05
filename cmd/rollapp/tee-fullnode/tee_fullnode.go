package teefullnode

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/consts"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/rollapp"
	"github.com/dymensionxyz/roller/utils/sequencer"
)

type TEEConfig struct {
	RollappID      string `toml:"rollapp_id"`
	HubRPCEndpoint string `toml:"hub_rpc_endpoint"`
	Home           string `toml:"home"`
}

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tee-fullnode <config-file>",
		Short: "Setup a TEE full node for validation",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			configFile := args[0]

			configFile, err := filesystem.ExpandHomePath(configFile)
			if err != nil {
				pterm.Error.Printf("expand home path: %v\n", err)
				return
			}

			var config TEEConfig
			_, err = toml.DecodeFile(configFile, &config)
			if err != nil {
				pterm.Error.Printf("parse config file: %v\n", err)
				return
			}

			if config.RollappID == "" {
				pterm.Error.Println("rollapp_id is required")
				return
			}

			if config.HubRPCEndpoint == "" {
				pterm.Error.Println("hub_rpc_endpoint is required")
				return
			}

			if config.Home == "" {
				config.Home = "/root/.roller"
			}

			config.Home, err = filesystem.ExpandHomePath(config.Home)
			if err != nil {
				pterm.Error.Printf("expand home directory: %v\n", err)
				return
			}

			hubData := consts.HubData{
				RpcUrl: config.HubRPCEndpoint,
				ID:     detectChainIDFromRPC(config.HubRPCEndpoint),
			}

			pterm.Info.Printf("Setting up TEE full node for %s\n", config.RollappID)
			pterm.Info.Printf("Home directory: %s\n", config.Home)
			pterm.Info.Printf("Hub RPC: %s\n", config.HubRPCEndpoint)

			rollappConfig, err := rollapp.PopulateRollerConfigWithRaMetadataFromChain(
				config.Home,
				config.RollappID,
				hubData,
			)
			if err != nil {
				pterm.Error.Printf("fetch rollapp metadata from chain: %v\n", err)
				return
			}

			raResponse, err := rollapp.GetMetadataFromChain(
				config.RollappID,
				hubData,
			)
			if err != nil {
				pterm.Error.Printf("fetch rollapp information from hub: %v\n", err)
				return
			}

			if !raResponse.Rollapp.GenesisInfo.Sealed {
				pterm.Error.Println("rollapp genesis is not sealed yet")
				return
			}

			pterm.Info.Println("Retrieving the latest available snapshot")
			si, err := sequencer.GetLatestSnapshot(
				rollappConfig.RollappID,
				rollappConfig.HubData,
			)
			if err != nil {
				pterm.Error.Printf("retrieve latest snapshot: %v\n", err)
				return
			}

			if si == nil {
				pterm.Error.Printf("no snapshots found for %s\n", rollappConfig.RollappID)
				return
			}

			pterm.Info.Printf("Found snapshot for height %s\n", si.Height)
			pterm.Info.Printf("Checksum: %s\n", si.Checksum)
			pterm.Info.Printf("URL: %s\n", si.SnapshotUrl)

			rollappDirPath := filepath.Join(config.Home, consts.ConfigDirName.Rollapp)
			dataDir := filepath.Join(rollappDirPath, "data")

			err = os.RemoveAll(dataDir)
			if err != nil {
				pterm.Error.Printf("remove data dir: %v\n", err)
				return
			}

			err = os.MkdirAll(dataDir, 0o755)
			if err != nil {
				pterm.Error.Printf("create data dir: %v\n", err)
				return
			}

			tmpDir, err := os.MkdirTemp("", "tee-snapshot-*")
			if err != nil {
				pterm.Error.Printf("create temp directory: %v\n", err)
				return
			}
			defer os.RemoveAll(tmpDir)

			archivePath := filepath.Join(tmpDir, "snapshot.tar.gz")
			spinner, _ := pterm.DefaultSpinner.Start("Downloading snapshot...")
			downloadedFileHash, err := filesystem.DownloadAndSaveArchive(
				si.SnapshotUrl,
				archivePath,
			)
			if err != nil {
				spinner.Fail(fmt.Sprintf("download snapshot: %v", err))
				return
			}
			spinner.Success("Snapshot downloaded successfully")

			if downloadedFileHash != si.Checksum {
				pterm.Error.Printf(
					"snapshot checksum mismatch, have: %s, want: %s\n",
					downloadedFileHash,
					si.Checksum,
				)
				return
			}

			err = filesystem.ExtractTarGz(archivePath, rollappDirPath)
			if err != nil {
				pterm.Error.Printf("extract snapshot: %v\n", err)
				return
			}

			peers, err := sequencer.GetAllP2pPeers(
				rollappConfig.RollappID,
				rollappConfig.HubData,
			)
			if err != nil {
				pterm.Warning.Printf("retrieve p2p peers: %v\n", err)
			}

			dymintFilePath := sequencer.GetDymintFilePath(config.Home)
			
			if len(peers) > 0 {
				peersStr := strings.Join(peers, ",")
				fieldsToUpdate := map[string]any{
					"p2p_bootstrap_nodes":  peersStr,
					"p2p_persistent_nodes": peersStr,
				}
				err = tomlconfig.UpdateFieldsInFile(dymintFilePath, fieldsToUpdate)
				if err != nil {
					pterm.Warning.Printf("add p2p peers: %v\n", err)
				}
			}

			damanager := datalayer.NewDAManager(
				rollappConfig.DA.Backend,
				rollappConfig.Home,
				rollappConfig.KeyringBackend,
				"fullnode",
			)

			daConfig := damanager.DataLayer.GetSequencerDAConfig("fullnode")
			
			vtu := map[string]any{
				"da_layer":                rollappConfig.DA.Backend,
				"da_config":               daConfig,
				"p2p_advertising_enabled": "false",
				"max_proof_time":          "5s",
				"settlement_gas_prices":   "20000000000adym",
			}

			err = tomlconfig.UpdateFieldsInFile(dymintFilePath, vtu)
			if err != nil {
				pterm.Error.Printf("update dymint config: %v\n", err)
				return
			}

			appConfigPath := sequencer.GetAppConfigFilePath(config.Home)
			appVtu := map[string]any{
				"api.enable":      false,
				"grpc.enable":     false,
				"json-rpc.enable": false,
			}
			
			err = tomlconfig.UpdateFieldsInFile(appConfigPath, appVtu)
			if err != nil {
				pterm.Error.Printf("update app config: %v\n", err)
				return
			}

			beConfigPath := filepath.Join(config.Home, consts.ConfigDirName.Rollapp, "config", "be-json-rpc.toml")
			err = tomlconfig.UpdateFieldInFile(beConfigPath, "enable", "false")
			if err != nil {
				pterm.Warning.Printf("disable block explorer: %v\n", err)
			}

			pterm.Success.Println("TEE full node setup complete")
			pterm.Info.Printf("To start the node, run: %s start --home %s\n", rollappConfig.RollappBinary, config.Home)
		},
	}

	return cmd
}

func detectChainIDFromRPC(rpcEndpoint string) string {
	if strings.Contains(rpcEndpoint, "playground") || 
	   strings.Contains(rpcEndpoint, "devnet") ||
	   strings.Contains(rpcEndpoint, "3405") {
		return "dymension_3405-1"
	}
	return "dymension_1100-1"
}