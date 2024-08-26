package utils

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type ContainerConfigOptions struct {
	Name   string
	Image  string
	Port   string
	Envs   []string
	Mounts []mount.Mount
}

// TODO: support multiple ports

// CreateContainer function create a mongo db docker container that is utilized by
// the eibc client to store fund information
func CreateContainer(
	ctx context.Context,
	cli *client.Client,
	cfg *ContainerConfigOptions,
) error {
	portString := fmt.Sprintf("%s/tcp", cfg.Port)

	portBindings := nat.PortMap{
		nat.Port(portString): []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: cfg.Port,
			},
		},
	}

	config := &container.Config{
		Image: cfg.Image,
		ExposedPorts: nat.PortSet{
			nat.Port(portString): struct{}{},
		},
		Env: []string{},
	}

	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
		Mounts:       cfg.Mounts,
		// Resources: container.Resources{
		// 	Memory: 256 * 1024 * 1024, // 512 MB
		// },
	}

	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return err
	}

	for _, c := range containers {
		for _, name := range c.Names {
			if strings.TrimPrefix(name, "/") == cfg.Name {
				fmt.Printf("Container %s already exists.\n", cfg.Name)

				if c.State != "running" {
					fmt.Printf(
						"Container %s is not in a running state, restarting.\n",
						cfg.Name,
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

	if !strings.HasPrefix(cfg.Image, "localhost") {
		pull, err := cli.ImagePull(ctx, cfg.Image, image.PullOptions{})
		if err != nil {
			return err
		}
		defer pull.Close()
	}

	resp, err := cli.ContainerCreate(ctx, config, hostConfig, nil, nil, cfg.Name)
	if err != nil {
		return err
	}

	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
			fmt.Printf("Failed to start container, attempt %d of %d: %v\n", i+1, maxRetries, err)
			time.Sleep(5 * time.Second)
			continue
		}

		logs, err := cli.ContainerLogs(
			ctx,
			resp.ID,
			container.LogsOptions{ShowStdout: true, ShowStderr: true},
		)
		if err != nil {
			return err
		}
		defer logs.Close()

		logContent, _ := io.ReadAll(logs)
		fmt.Println("Container logs:", string(logContent))

		fmt.Printf("Container %s created and started successfully\n", cfg.Name)
		return nil
	}

	return fmt.Errorf("failed to start container after %d attempts", maxRetries)
}
