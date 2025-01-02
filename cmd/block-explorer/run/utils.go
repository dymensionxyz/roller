package run

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	_ "github.com/lib/pq"
	"github.com/pterm/pterm"
	"golang.org/x/exp/maps"

	"github.com/dymensionxyz/roller/cmd/consts"
	postgresqlutils "github.com/dymensionxyz/roller/utils/database/postgresql"
	"github.com/dymensionxyz/roller/utils/dependencies/types"
	dockerutils "github.com/dymensionxyz/roller/utils/docker"
	"github.com/dymensionxyz/roller/utils/filesystem"
)

func createBlockExplorerContainers(home, hostAddress string) error {
	pterm.Info.Println("Creating container for block explorer")
	cc, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		fmt.Printf("Failed to create Docker client: %v\n", err)
		return err
	}

	parsedHostAddress, err := url.Parse(hostAddress)
	if err != nil {
		fmt.Println("Error parsing URL:", err)
		return err
	}

	domain := parsedHostAddress.Hostname()

	networkName := "block_explorer_network"
	if err := ensureNetworkExists(cc, networkName); err != nil {
		fmt.Printf("Failed to ensure network: %v\n", err)
		return err
	}

	beChainConfigPath := filepath.Join(
		home,
		consts.ConfigDirName.BlockExplorer,
		"config",
		"chains.yaml",
	)
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
			Image: "public.ecr.aws/a3d4b9r3/block-explorer-frontend:next",
			Port:  "3000",
			Envs: []string{
				fmt.Sprintf("DATABASE_URL=postgresql://be:psw@%s:5432/blockexplorer", domain),
				fmt.Sprintf("HOST_ADDRESS=%s", domain),
			},
			Mounts: []mount.Mount{},
		},
		"indexer": {
			Name:  "be-indexer",
			Image: "public.ecr.aws/a3d4b9r3/block-explorer-indexer:next",
			Port:  "8080",
			Envs: []string{
				fmt.Sprintf("HOST_ADDRESS=%s", domain),
			},
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: beChainConfigPath,
					Target: "/root/.beid/chains.yaml",
				},
			},
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

		// Connect the container to the network using the new function
		err = connectContainerToNetwork(context.Background(), cc, networkName, options.Name)
		if err != nil {
			fmt.Printf("Error with network connection for container %s: %v\n", options.Name, err)
			return err
		}
	}

	if err := runSQLMigration(home); err != nil {
		fmt.Printf("Failed to apply migrations: %v\n", err)
		return err
	}

	return nil
}

func ensureNetworkExists(cli *client.Client, networkName string) error {
	// List all networks
	networks, err := cli.NetworkList(context.Background(), network.ListOptions{})
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
		context.Background(), networkName, network.CreateOptions{
			Driver: "bridge",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create network: %w", err)
	}

	fmt.Printf("Network %s created successfully.\n", networkName)
	return nil
}

func connectContainerToNetwork(
	ctx context.Context,
	cc *client.Client,
	networkName, containerName string,
) error {
	networkResource, err := cc.NetworkInspect(ctx, networkName, network.InspectOptions{})
	if err != nil {
		return fmt.Errorf("failed to inspect network: %v", err)
	}

	for _, container := range networkResource.Containers {
		if container.Name == containerName {
			fmt.Printf(
				"Container %s is already connected to network %s\n",
				containerName,
				networkName,
			)
			return nil
		}
	}

	err = cc.NetworkConnect(
		ctx,
		networkName,
		containerName,
		&network.EndpointSettings{},
	)
	if err != nil {
		return fmt.Errorf("failed to connect container %s to network: %v", containerName, err)
	}

	fmt.Printf("Connected container %s to network %s\n", containerName, networkName)
	return nil
}

func runSQLMigration(home string) error {
	dbHost := "localhost"
	dbPort := "5432"
	dbName := "blockexplorer"
	dbUserAdmin := "be"
	dbPassAdmin := "psw"

	dbConnStr := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		dbUserAdmin,
		dbPassAdmin,
		dbHost,
		dbPort,
		dbName,
	)

	pterm.Info.Println("Retrieving SQL migration files")
	dbMigrationsPath := filepath.Join(home, consts.ConfigDirName.BlockExplorer, "migrations")
	dbMigrationsSchemaPath := filepath.Join(dbMigrationsPath, "schema.sql")
	dbMigrationsEventsPath := filepath.Join(dbMigrationsPath, "events.sql")

	err := os.MkdirAll(dbMigrationsPath, 0o755)
	if err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}

	migrationFiles := []types.PersistFile{
		{
			Source: "https://raw.githubusercontent.com/dymensionxyz/roller/main/migrations/block-explorer/schema.sql",
			Target: dbMigrationsSchemaPath,
		},
		{
			Source: "https://raw.githubusercontent.com/dymensionxyz/roller/main/migrations/block-explorer/events.sql",
			Target: dbMigrationsEventsPath,
		},
	}

	for _, file := range migrationFiles {
		err := filesystem.DownloadFile(file.Source, file.Target)
		if err != nil {
			pterm.Error.Printf("Failed to retrieve SQL migration %s: %v\n", file.Target, err)
			return err
		}
	}

	dbAdmin, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		return fmt.Errorf("failed to connect to database as admin: %w", err)
	}
	defer dbAdmin.Close()

	dbLocal, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		return fmt.Errorf("failed to connect to database as local user: %w", err)
	}
	defer dbLocal.Close()

	// Create migration tracking table
	const createMigrationTableSQL = `
    CREATE TABLE IF NOT EXISTS applied_migrations (
        id SERIAL PRIMARY KEY,
        filename VARCHAR(255) NOT NULL UNIQUE,
        applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );`

	_, err = dbAdmin.Exec(createMigrationTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create migration table: %w", err)
	}

	// Apply schema migration
	schemaContent, err := os.ReadFile(dbMigrationsSchemaPath)
	if err != nil {
		return fmt.Errorf("failed to read schema SQL file: %w", err)
	}
	err = postgresqlutils.ApplyMigration(dbLocal, "schema.sql", schemaContent)
	if err != nil {
		return err
	}

	// Apply events migration
	eventsContent, err := os.ReadFile(dbMigrationsEventsPath)
	if err != nil {
		return fmt.Errorf("failed to read events SQL file: %w", err)
	}
	err = postgresqlutils.ApplyMigration(dbAdmin, "events.sql", eventsContent)
	if err != nil {
		return err
	}

	pterm.Success.Println("Migrations checked and applied successfully")
	return nil
}
