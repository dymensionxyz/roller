package teefullnode

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// TEEConfig represents the configuration for TEE full node
type TEEConfig struct {
	// Required fields
	RollappID      string `yaml:"rollapp_id" json:"rollapp_id"`
	HubRPCEndpoint string `yaml:"hub_rpc_endpoint" json:"hub_rpc_endpoint"`
	DARPCRPC       string `yaml:"da_rpc_endpoint" json:"da_rpc_endpoint"`
	
	// Optional fields
	Home                string `yaml:"home,omitempty" json:"home,omitempty"`
	SnapshotURL         string `yaml:"snapshot_url,omitempty" json:"snapshot_url,omitempty"`
	BatchSubmitInterval string `yaml:"batch_submit_interval,omitempty" json:"batch_submit_interval,omitempty"`
	RPCPort             string `yaml:"rpc_port,omitempty" json:"rpc_port,omitempty"`
	APIPort             string `yaml:"api_port,omitempty" json:"api_port,omitempty"`
	P2PPort             string `yaml:"p2p_port,omitempty" json:"p2p_port,omitempty"`
	LogLevel            string `yaml:"log_level,omitempty" json:"log_level,omitempty"`
	SkipDALightClient   bool   `yaml:"skip_da_light_client" json:"skip_da_light_client"`
}

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tee-fullnode",
		Short: "Start a TEE full node with non-interactive configuration",
		Long: `Start a TEE full node using configuration from file or environment.

This command is designed for running inside TEE environments where interactive
prompts are not available. All configuration must be provided via:
- Config file (--config flag)
- Command line flags
- Environment variables (prefixed with TEE_)

Example config file (tee-config.yaml):
  rollapp_id: "myrollapp_1234-1"
  hub_rpc_endpoint: "https://hub-rpc.example.com"
  da_rpc_endpoint: "https://da-rpc.example.com"
  snapshot_url: "https://snapshot.example.com/latest.tar.gz"
  batch_submit_interval: "100"
  skip_da_light_client: true
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			configFile, _ := cmd.Flags().GetString("config")
			
			// Load configuration
			config, err := loadTEEConfig(cmd, configFile)
			if err != nil {
				return fmt.Errorf("failed to load TEE config: %w", err)
			}
			
			// Validate required fields
			if err := validateTEEConfig(config); err != nil {
				return fmt.Errorf("invalid TEE config: %w", err)
			}
			
			// Set defaults
			if config.Home == "" {
				config.Home = filepath.Join(os.Getenv("HOME"), ".roller", "rollapp")
			}
			if config.RPCPort == "" {
				config.RPCPort = "26657"
			}
			if config.APIPort == "" {
				config.APIPort = "1317"
			}
			if config.P2PPort == "" {
				config.P2PPort = "26656"
			}
			if config.LogLevel == "" {
				config.LogLevel = "info"
			}
			
			// Initialize or restore from snapshot
			if config.SnapshotURL != "" {
				pterm.Info.Printf("Restoring from snapshot: %s\n", config.SnapshotURL)
				if err := restoreFromSnapshot(config); err != nil {
					return fmt.Errorf("failed to restore from snapshot: %w", err)
				}
			} else {
				pterm.Info.Println("Initializing full node from genesis")
				if err := initializeFullNode(config); err != nil {
					return fmt.Errorf("failed to initialize full node: %w", err)
				}
			}
			
			// Configure the node for TEE mode
			if err := configureTEENode(config); err != nil {
				return fmt.Errorf("failed to configure TEE node: %w", err)
			}
			
			// Start the full node
			pterm.Success.Println("Starting TEE full node...")
			return startFullNode(config)
		},
	}
	
	// Add flags
	cmd.Flags().String("config", "", "Path to TEE configuration file (YAML or JSON)")
	cmd.Flags().String("rollapp-id", "", "Rollapp ID")
	cmd.Flags().String("hub-rpc", "", "Hub RPC endpoint")
	cmd.Flags().String("da-rpc", "", "DA RPC endpoint")
	cmd.Flags().String("snapshot-url", "", "Snapshot URL for restoration")
	cmd.Flags().String("batch-submit-interval", "", "Batch submission interval")
	cmd.Flags().String("home", "", "Home directory for the node")
	cmd.Flags().Bool("skip-da-light-client", false, "Skip running local DA light client")
	
	return cmd
}

func loadTEEConfig(cmd *cobra.Command, configFile string) (*TEEConfig, error) {
	config := &TEEConfig{}
	
	// Load from file if provided
	if configFile != "" {
		data, err := os.ReadFile(configFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		
		// Try JSON first, then YAML
		if err := json.Unmarshal(data, config); err != nil {
			if err := yaml.Unmarshal(data, config); err != nil {
				return nil, fmt.Errorf("failed to parse config file as JSON or YAML: %w", err)
			}
		}
	}
	
	// Override with command line flags
	if val, _ := cmd.Flags().GetString("rollapp-id"); val != "" {
		config.RollappID = val
	}
	if val, _ := cmd.Flags().GetString("hub-rpc"); val != "" {
		config.HubRPCEndpoint = val
	}
	if val, _ := cmd.Flags().GetString("da-rpc"); val != "" {
		config.DARPCRPC = val
	}
	if val, _ := cmd.Flags().GetString("snapshot-url"); val != "" {
		config.SnapshotURL = val
	}
	if val, _ := cmd.Flags().GetString("batch-submit-interval"); val != "" {
		config.BatchSubmitInterval = val
	}
	if val, _ := cmd.Flags().GetString("home"); val != "" {
		config.Home = val
	}
	if val, _ := cmd.Flags().GetBool("skip-da-light-client"); val {
		config.SkipDALightClient = val
	}
	
	// Override with environment variables
	if val := os.Getenv("TEE_ROLLAPP_ID"); val != "" {
		config.RollappID = val
	}
	if val := os.Getenv("TEE_HUB_RPC"); val != "" {
		config.HubRPCEndpoint = val
	}
	if val := os.Getenv("TEE_DA_RPC"); val != "" {
		config.DARPCRPC = val
	}
	if val := os.Getenv("TEE_SNAPSHOT_URL"); val != "" {
		config.SnapshotURL = val
	}
	if val := os.Getenv("TEE_BATCH_SUBMIT_INTERVAL"); val != "" {
		config.BatchSubmitInterval = val
	}
	if val := os.Getenv("TEE_HOME"); val != "" {
		config.Home = val
	}
	if val := os.Getenv("TEE_SKIP_DA_LIGHT_CLIENT"); val == "true" {
		config.SkipDALightClient = true
	}
	
	return config, nil
}

func validateTEEConfig(config *TEEConfig) error {
	if config.RollappID == "" {
		return fmt.Errorf("rollapp_id is required")
	}
	if config.HubRPCEndpoint == "" {
		return fmt.Errorf("hub_rpc_endpoint is required")
	}
	if config.DARPCRPC == "" && !config.SkipDALightClient {
		return fmt.Errorf("da_rpc_endpoint is required when not skipping DA light client")
	}
	return nil
}

func restoreFromSnapshot(config *TEEConfig) error {
	// Create home directory if it doesn't exist
	if err := os.MkdirAll(config.Home, 0755); err != nil {
		return fmt.Errorf("failed to create home directory: %w", err)
	}
	
	// Download and extract snapshot
	pterm.Info.Printf("Downloading snapshot from %s\n", config.SnapshotURL)
	
	// Use wget or curl to download
	tmpFile := filepath.Join("/tmp", "snapshot.tar.gz")
	downloadCmd := exec.Command("wget", "-O", tmpFile, config.SnapshotURL)
	if err := downloadCmd.Run(); err != nil {
		// Try curl as fallback
		downloadCmd = exec.Command("curl", "-L", "-o", tmpFile, config.SnapshotURL)
		if err := downloadCmd.Run(); err != nil {
			return fmt.Errorf("failed to download snapshot: %w", err)
		}
	}
	
	// Extract snapshot
	pterm.Info.Println("Extracting snapshot...")
	extractCmd := exec.Command("tar", "-xzf", tmpFile, "-C", config.Home)
	if err := extractCmd.Run(); err != nil {
		return fmt.Errorf("failed to extract snapshot: %w", err)
	}
	
	// Clean up
	os.Remove(tmpFile)
	
	return nil
}

func initializeFullNode(config *TEEConfig) error {
	// Initialize the node using roller init logic
	// This is a simplified version - in production, use the actual roller init logic
	
	pterm.Info.Printf("Initializing full node for rollapp: %s\n", config.RollappID)
	
	// Create necessary directories
	dirs := []string{
		filepath.Join(config.Home, "config"),
		filepath.Join(config.Home, "data"),
		filepath.Join(config.Home, "keyring-test"),
	}
	
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	
	// TODO: Initialize genesis, config files, etc. based on rollapp ID
	// This would typically involve:
	// 1. Fetching genesis from hub
	// 2. Configuring app.toml and config.toml
	// 3. Setting up peers and seeds
	
	return nil
}

func configureTEENode(config *TEEConfig) error {
	pterm.Info.Println("Configuring node for TEE mode...")
	
	// Update config.toml
	configPath := filepath.Join(config.Home, "config", "config.toml")
	appConfigPath := filepath.Join(config.Home, "config", "app.toml")
	
	// Set RPC endpoints
	updateConfig := []string{
		fmt.Sprintf("s|^laddr = \"tcp://.*:26657\"|laddr = \"tcp://0.0.0.0:%s\"|", config.RPCPort),
		fmt.Sprintf("s|^laddr = \"tcp://.*:26656\"|laddr = \"tcp://0.0.0.0:%s\"|", config.P2PPort),
	}
	
	for _, sed := range updateConfig {
		cmd := exec.Command("sed", "-i", sed, configPath)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to update config.toml: %w", err)
		}
	}
	
	// Update app.toml for API
	updateAppConfig := []string{
		fmt.Sprintf("s|address = \"tcp://.*:1317\"|address = \"tcp://0.0.0.0:%s\"|", config.APIPort),
		"s|enable = false|enable = true|",
	}
	
	for _, sed := range updateAppConfig {
		cmd := exec.Command("sed", "-i", sed, appConfigPath)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to update app.toml: %w", err)
		}
	}
	
	// Set batch submission interval if provided
	if config.BatchSubmitInterval != "" {
		// TODO: Update dymint config with batch submission interval
		pterm.Info.Printf("Setting batch submission interval to: %s\n", config.BatchSubmitInterval)
	}
	
	// Configure external DA RPC if provided
	if config.DARPCRPC != "" && config.SkipDALightClient {
		// TODO: Configure to use external DA RPC instead of local light client
		pterm.Info.Printf("Using external DA RPC: %s\n", config.DARPCRPC)
	}
	
	// Configure Hub RPC
	// TODO: Update hub connection configuration
	pterm.Info.Printf("Configuring Hub RPC: %s\n", config.HubRPCEndpoint)
	
	return nil
}

func startFullNode(config *TEEConfig) error {
	// Determine the rollapp binary based on the rollapp ID
	// In a real implementation, this would be more sophisticated
	rollappBinary := "rollapp-evm" // Default, should be determined from config
	
	// Build the command
	args := []string{
		"start",
		"--home", config.Home,
		"--log-level", config.LogLevel,
	}
	
	// Add TEE attestation endpoint flag if needed
	args = append(args, "--tee-attestation-enabled")
	
	pterm.Info.Printf("Starting rollapp with command: %s %s\n", rollappBinary, strings.Join(args, " "))
	
	// Start the node
	cmd := exec.Command(rollappBinary, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	
	// Set environment variables
	cmd.Env = os.Environ()
	if config.HubRPCEndpoint != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("HUB_RPC=%s", config.HubRPCEndpoint))
	}
	if config.DARPCRPC != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("DA_RPC=%s", config.DARPCRPC))
	}
	
	return cmd.Run()
}