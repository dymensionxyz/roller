package teardown

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "teardown",
		Short: "Remove block explorer resources",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			err := removeBlockExplorerResources()
			if err != nil {
				pterm.Error.Println("failed to remove Block Explorer resources: ", err)
				return
			}
		},
	}

	return cmd
}

func removeBlockExplorerResources() error {
	spinner, _ := pterm.DefaultSpinner.Start("Removing Block Explorer resources")

	ctx := context.Background()
	cc, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %v", err)
	}
	defer cc.Close()

	// Container names
	containerNames := []string{"be-postgresql", "be-frontend", "be-indexer"}

	// Remove containers
	for _, name := range containerNames {
		pterm.Info.Printf("Removing container: %s\n", name)
		err := cc.ContainerRemove(ctx, name, container.RemoveOptions{Force: true})
		if err != nil && !client.IsErrNotFound(err) {
			pterm.Warning.Printf("Failed to remove container %s: %v\n", name, err)
		}
	}

	// Remove network
	networkName := "block_explorer_network"
	pterm.Info.Printf("Removing network: %s\n", networkName)
	err = cc.NetworkRemove(ctx, networkName)
	if err != nil && !client.IsErrNotFound(err) {
		pterm.Warning.Printf("Failed to remove network %s: %v\n", networkName, err)
	}

	// Remove volume
	volumeName := "postgres_data"
	pterm.Info.Printf("Removing volume: %s\n", volumeName)
	err = cc.VolumeRemove(ctx, volumeName, true)
	if err != nil && !client.IsErrNotFound(err) {
		pterm.Warning.Printf("Failed to remove volume %s: %v\n", volumeName, err)
	}

	spinner.Success("Block Explorer resources removed successfully")
	return nil
}
