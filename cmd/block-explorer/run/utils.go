package run

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	dockerutils "github.com/dymensionxyz/roller/utils/docker"
	"github.com/pterm/pterm"
	"golang.org/x/exp/maps"
)

func createBlockExplorerContainers() error {
	pterm.Info.Println("creating container for block explorer")
	cc, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		fmt.Printf("failed to create docker client: %v\n", err)
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
			Name:   "be-frontend",
			Image:  "localhost/block-explorer:latest",
			Port:   "3000",
			Envs:   []string{},
			Mounts: []mount.Mount{},
		},
		// "indexer": {
		// 	Name:  "be-indexer",
		// 	Image: "postgres:16-alpine",
		// 	Port:  "5432",
		// },
	}

	pterm.Info.Printf("that will be created: %s\n", strings.Join(maps.Keys(containers), ", "))

	for _, options := range containers {
		err = dockerutils.CreateContainer(
			context.Background(),
			cc,
			&options,
		)
		if err != nil {
			fmt.Printf("failed to run %s container: %v\n", options.Name, err)
			return err
		}
		return err
	}

	return nil
}
