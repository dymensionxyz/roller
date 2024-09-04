package run

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	_ "github.com/lib/pq"
	"github.com/pterm/pterm"
	"golang.org/x/exp/maps"

	dockerutils "github.com/dymensionxyz/roller/utils/docker"
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

	pterm.Info.Printf(
		"Containers that will be created: %s\n",
		strings.Join(maps.Keys(containers), ", "),
	)

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
		err = cc.NetworkConnect(
			context.Background(),
			networkName,
			options.Name,
			&network.EndpointSettings{},
		)
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
	// Database connection details
	dbHost := "localhost"
	dbPort := "5432"
	dbName := "blockexplorer"
	dbUserAdmin := "be"
	dbPassAdmin := "psw"

	// Connect to the database as an admin user
	dbConnStr := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		dbUserAdmin,
		dbPassAdmin,
		dbHost,
		dbPort,
		dbName,
	)
	dbAdmin, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		return fmt.Errorf("failed to connect to database as admin: %w", err)
	}
	defer dbAdmin.Close()

	// Wait for the database to be ready
	time.Sleep(5 * time.Second)

	// Execute PostgreSQL commands to set up the database and roles
	// setupCommands := []string{
	// 	fmt.Sprintf("ALTER ROLE %s WITH LOGIN;", dbUserAdmin),
	// 	fmt.Sprintf("GRANT ALL PRIVILEGES ON DATABASE %s TO %s;", dbName, dbUserAdmin),
	// 	fmt.Sprintf("GRANT ALL PRIVILEGES ON SCHEMA public TO %s;", dbUserAdmin),
	// }

	// for _, cmd := range setupComman
	// 	_, err = dbAdmin.Exec(cmd)
	// 	if err != nil {
	// 		return fmt.Errorf("failed to execute setup command: %w", err)
	// 	}
	// }

	// Connect to the new database as the local user
	dbLocal, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		return fmt.Errorf("failed to connect to database as local user: %w", err)
	}
	defer dbLocal.Close()

	// Read and execute the SQL migration file
	sqlFile, err := os.ReadFile("migrations/blockexplorer.sql")
	if err != nil {
		return fmt.Errorf("failed to read SQL file: %w", err)
	}

	_, err = dbLocal.Exec(string(sqlFile))
	if err != nil {
		return fmt.Errorf("failed to execute SQL migration: %w", err)
	}

	// Execute additional SQL files if needed
	superSchemaFile, err := os.ReadFile("migrations/super-schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read super-schema SQL file: %w", err)
	}

	_, err = dbAdmin.Exec(string(superSchemaFile))
	if err != nil {
		return fmt.Errorf("failed to execute super-schema SQL file: %w", err)
	}

	log.Println("Migrations applied successfully")
	return nil
}
