package run

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	dockerutils "github.com/dymensionxyz/roller/utils/docker"
	"github.com/pterm/pterm"
	"golang.org/x/exp/maps"
)

func createBlockExplorerContainers() error {
	pterm.Info.Println("Creating container for block explorer")
	cc, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		fmt.Printf("Failed to create Docker client: %v\n", err)
		return err
	}

	networkName := "block_explorer_network"
	if err := ensureNetworkExists(cc, networkName); err != nil {
		fmt.Printf("Failed to ensure network: %v\n", err)
		return err
	}

	containers := map[string]dockerutils.ContainerConfigOptions{
		"db": {
			Name:  "be-postgresql",
			Image: "postgres:16-alpine",
			Port:  "5432",
			Envs: []string{
				"POSTGRES_USER=be",
				"POSTGRES_PASSWORD=psw",
				"POSTGRES_DB=blockexplorer",
			},
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeVolume,
					Source: "postgres_data",
					Target: "/var/lib/postgresql/data",
				},
			},
		},
		"frontend": {
			Name:  "be-frontend",
			Image: "localhost/block-explorer:latest",
			Port:  "3000",
			Envs: []string{
				"DATABASE_URL=postgresql://be:psw@be-postgresql:5432/blockexplorer",
			},
			Mounts: []mount.Mount{},
		},
		"indexer": {
			Name:   "be-indexer",
			Image:  "localhost/block-explorer-indexer:latest",
			Port:   "8080",
			Envs:   []string{},
			Mounts: []mount.Mount{},
		},
	}

	pterm.Info.Printf("Containers that will be created: %s\n", strings.Join(maps.Keys(containers), ", "))

	for _, options := range containers {
		err = dockerutils.CreateContainer(
			context.Background(),
			cc,
			&options,
		)
		if err != nil {
			fmt.Printf("Failed to run %s container: %v\n", options.Name, err)
			return err
		}

		// Connect the container to the network
		err = cc.NetworkConnect(context.Background(), networkName, options.Name, &network.EndpointSettings{})
		if err != nil {
			fmt.Printf("Failed to connect container %s to network: %v\n", options.Name, err)
			return err
		}
	}

	if err := runSQLMigration(); err != nil {
		fmt.Printf("Failed to apply migrations: %v\n", err)
		return err
	}

	return nil
}

func ensureNetworkExists(cli *client.Client, networkName string) error {
	// List all networks
	networks, err := cli.NetworkList(context.Background(), types.NetworkListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list networks: %w", err)
	}

	// Check if the network already exists
	for _, network := range networks {
		if network.Name == networkName {
			fmt.Printf("Network %s already exists, skipping creation.\n", networkName)
			return nil
		}
	}

	// Create the network if it does not exist
	_, err = cli.NetworkCreate(
		context.Background(), networkName, types.NetworkCreate{
			Driver: "bridge",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create network: %w", err)
	}

	fmt.Printf("Network %s created successfully.\n", networkName)
	return nil
}

func runSQLMigration() error {
	// Read the SQL file
	sqlFile, err := ioutil.ReadFile("migrations/blockexplorer.sql")
	if err != nil {
		return fmt.Errorf("failed to read SQL file: %w", err)
	}

	// Connect to the database
	connStr := "postgresql://be:psw@localhost:5432/blockexplorer?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Split the SQL script into individual statements
	// Note: This is a simple split and may not handle complex SQL with semicolons inside strings or comments.
	queries := strings.Split(string(sqlFile), ";")

	// Execute each SQL statement
	for _, query := range queries {
		query = strings.TrimSpace(query)
		if query == "" {
			continue
		}
		_, err := db.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to execute SQL statement: %w", err)
		}
	}

	log.Println("Migrations applied successfully")
	return nil
}
