package tasks

import (
	"context"
	"io"
	"log"
	"math"
	"os"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"golang.org/x/exp/slices"
)

type State int

const (
	Pending State = iota
	Scheduled
	Running
	Completed
	Failed
)

var stateTransitionMap = map[State][]State{
	Pending:   {Scheduled},
	Scheduled: {Scheduled, Running, Failed},
	Running:   {Running, Completed, Failed},
	Completed: {},
	Failed:    {},
}

func ValidStateTransition(src State, dst State) bool {
	validDst := stateTransitionMap[src]
	return slices.Contains(validDst, dst)
}

type Task struct {
	ID    uuid.UUID
	Name  string
	Image string

	State       State
	ContainerID string

	StartTime  time.Time
	FinishTime time.Time
}

type TaskRequest struct {
	ID            uuid.UUID
	Name          string
	Image         string
	RequiredState State
}

func (t TaskRequest) Task() Task {
	return Task{
		ID:    t.ID,
		Name:  t.Name,
		Image: t.Image,
		State: t.RequiredState,
	}
}

func (t Task) Config() Config {
	return Config{
		Name:  t.Name,
		Image: t.Image,
	}
}

type TaskEvent struct {
	ID        uuid.UUID
	State     State
	Timestamp time.Time
	Task      Task
}

type Config struct {
	Name          string
	AttachStdin   bool
	AttachStdout  bool
	AttachStderr  bool
	ExposedPorts  nat.PortSet
	Cmd           []string
	Image         string
	Cpu           float64
	Memory        int64
	Disk          int64
	Env           []string
	RestartPolicy container.RestartPolicyMode

	Runtime Runtime
}

type Runtime struct {
	ContainerID string
}

type Docker struct {
	Client *client.Client
}

type DockerResult struct {
	Error       error
	Action      string
	ContainerId string
	Result      string
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

	out, err := d.Client.ContainerLogs(ctx, createResponse.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		log.Printf("Error getting logs for container %s: %v\n", createResponse.ID, err)
		return DockerResult{Error: err}
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	return DockerResult{ContainerId: createResponse.ID, Action: "start", Result: "success"}
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

	return DockerResult{Action: "stop", Result: "success", Error: nil}
}
