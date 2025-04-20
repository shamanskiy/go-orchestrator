package tasks

import (
	"context"
	"io"
	"log"
	"math"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type Docker struct {
	Client *client.Client
}

type DockerResult struct {
	ContainerId string
	Port        int
	Error       error
}

func (d *Docker) Run(config Config) DockerResult {
	ctx := context.Background()
	reader, err := d.Client.ImagePull(ctx, config.Image, image.PullOptions{})
	if err != nil {
		log.Printf("Error pulling image %s: %v\n", config.Image, err)
		return DockerResult{Error: err}
	}
	io.Copy(os.Stdout, reader)

	restartPolicy := container.RestartPolicy{
		Name: config.RestartPolicy,
	}
	resources := container.Resources{
		Memory:   config.Memory,
		NanoCPUs: int64(config.Cpu * math.Pow(10, 9)),
	}
	containerConfig := container.Config{
		Image:        config.Image,
		Tty:          false,
		Env:          config.Env,
		ExposedPorts: config.ExposedPorts,
	}
	hostConfig := container.HostConfig{
		RestartPolicy:   restartPolicy,
		Resources:       resources,
		PublishAllPorts: true,
	}

	createResponse, err := d.Client.ContainerCreate(ctx, &containerConfig, &hostConfig, nil, nil, config.Name)
	if err != nil {
		log.Printf("Error creating container using image %s: %v\n", config.Image, err)
		return DockerResult{Error: err}
	}

	err = d.Client.ContainerStart(ctx, createResponse.ID, container.StartOptions{})
	if err != nil {
		log.Printf("Error starting container %s: %v\n", createResponse.ID, err)
		return DockerResult{Error: err}
	}

	inspectResponse, err := d.Client.ContainerInspect(ctx, createResponse.ID)
	if err != nil {
		log.Printf("Error inspecting container %s: %v\n", createResponse.ID, err)
		return DockerResult{Error: err}
	}
	portBindings := inspectResponse.NetworkSettings.Ports

	portStr := portBindings[nat.Port("80/tcp")][0].HostPort
	port, err := nat.ParsePort(portStr)
	if err != nil {
		log.Printf("Error parsing port %s: %v\n", portStr, err)
		return DockerResult{Error: err}
	}

	return DockerResult{ContainerId: createResponse.ID, Port: port}
}

func (d *Docker) Remove(containerId string) DockerResult {
	log.Printf("Attempting to stop container %v", containerId)
	ctx := context.Background()
	err := d.Client.ContainerStop(ctx, containerId, container.StopOptions{})
	if err != nil {
		log.Printf("Error stopping container %s: %v\n", containerId, err)
		return DockerResult{Error: err}
	}

	err = d.Client.ContainerRemove(ctx, containerId, container.RemoveOptions{
		RemoveVolumes: true,
		RemoveLinks:   false,
		Force:         false,
	})
	if err != nil {
		log.Printf("Error removing container %s: %v\n", containerId, err)
		return DockerResult{Error: err}
	}

	return DockerResult{}
}
