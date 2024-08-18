package utils

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// TODO: this could be a generic docker container helper

// CheckAndCreateMongoDBContainer function create a mongo db docker container that is utilized by
// the eibc client to store fund information
func CheckAndCreateMongoDBContainer(
	ctx context.Context,
	cli *client.Client,
) error {
	containerName := "eibc-mongodb"
	imageName := "mongo:7.0"

	portBindings := nat.PortMap{
		"27017/tcp": []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: "27017",
			},
		},
	}

	config := &container.Config{
		Image: imageName,
		ExposedPorts: nat.PortSet{
			"27017/tcp": struct{}{},
		},
	}

	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
	}

	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return err
	}

	for _, c := range containers {
		for _, name := range c.Names {
			if strings.TrimPrefix(name, "/") == containerName {
				fmt.Printf("Container %s already exists.\n", containerName)

				if c.State != "running" {
					fmt.Printf(
						"Container %s is not in a running state, restarting.\n",
						containerName,
					)
					err := cli.ContainerStart(context.Background(), c.ID, container.StartOptions{})
					if err != nil {
						return err
					}
				}

				return nil
			}
		}
	}

	pull, err := cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return err
	}
	defer pull.Close()

	resp, err := cli.ContainerCreate(ctx, config, hostConfig, nil, nil, containerName)
	if err != nil {
		return err
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return err
	}

	fmt.Printf("Container %s created and started successfully\n", containerName)
	return nil
}
