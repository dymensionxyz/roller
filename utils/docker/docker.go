package docker

import (
	"context"
	"fmt"
	"io"
	"log"
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
		Env: cfg.Envs,
	}

	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
		Mounts:       cfg.Mounts,
	}

	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return fmt.Errorf("error listing containers: %w", err)
	}

	for _, c := range containers {
		for _, name := range c.Names {
			if strings.TrimPrefix(name, "/") == cfg.Name {
				log.Printf("Container %s already exists.\n", cfg.Name)

				if c.State != "running" {
					log.Printf("Container %s is not running, restarting.\n", cfg.Name)
					if err := cli.ContainerStart(ctx, c.ID, container.StartOptions{}); err != nil {
						return fmt.Errorf("error starting container %s: %w", cfg.Name, err)
					}
				}

				return nil
			}
		}
	}

	if !strings.HasPrefix(cfg.Image, "localhost") {
		pull, err := cli.ImagePull(ctx, cfg.Image, image.PullOptions{})
		if err != nil {
			return fmt.Errorf("error pulling image %s: %w", cfg.Image, err)
		}
		defer pull.Close()
		io.Copy(io.Discard, pull) // Ensure the pull stream is fully read
	}

	resp, err := cli.ContainerCreate(ctx, config, hostConfig, nil, nil, cfg.Name)
	if err != nil {
		return fmt.Errorf("error creating container %s: %w", cfg.Name, err)
	}

	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
			log.Printf("Failed to start container, attempt %d of %d: %v\n", i+1, maxRetries, err)
			time.Sleep(5 * time.Second)
			continue
		}

		logs, err := cli.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
		if err != nil {
			return fmt.Errorf("error retrieving logs for container %s: %w", cfg.Name, err)
		}
		defer logs.Close()

		logContent, _ := io.ReadAll(logs)
		log.Println("Container logs:", string(logContent))

		log.Printf("Container %s created and started successfully\n", cfg.Name)
		return nil
	}

	return fmt.Errorf("failed to start container after %d attempts", maxRetries)
}
