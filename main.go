package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/docker/docker/client"
	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
	"github.com/shamanskiy/go-orchestrator/managers"
	"github.com/shamanskiy/go-orchestrator/nodes"
	"github.com/shamanskiy/go-orchestrator/tasks"
	"github.com/shamanskiy/go-orchestrator/workers"
)

func main() {
	task := tasks.Task{
		ID:     uuid.New(),
		Name:   "Task-1",
		State:  tasks.Pending,
		Image:  "Image-1",
		Memory: 1024,
		Disk:   1}

	taskEvent := tasks.TaskEvent{
		ID:        uuid.New(),
		State:     tasks.Pending,
		Timestamp: time.Now(),
		Task:      task,
	}
	fmt.Printf("task: %v\n", task)
	fmt.Printf("task event: %v\n", taskEvent)

	worker := workers.Worker{
		Name:  "worker-1",
		Queue: *queue.New(),
		Db:    make(map[uuid.UUID]*tasks.Task),
	}

	fmt.Printf("worker: %v\n", worker)
	worker.CollectStats()
	worker.RunTask()
	worker.StartTask()
	worker.StopTask()

	manager := managers.Manager{
		Pending: *queue.New(),
		TaskDb:  make(map[string][]tasks.Task),
		EventDb: make(map[string][]tasks.TaskEvent),
		Workers: []string{worker.Name},
	}

	fmt.Printf("manager: %v\n", manager)
	manager.SelectWorker()
	manager.UpdateTasks()
	manager.SendWork()

	node := nodes.Node{
		Name:   "Node-1",
		Ip:     "192.168.1.1",
		Cores:  4,
		Memory: 1024,
		Disk:   25,
		Role:   "worker",
	}
	fmt.Printf("node: %v\n", node)

	dockerTask, createResult := createContainer()
	if createResult.Error != nil {
		fmt.Printf("%v", createResult.Error)
		os.Exit(1)
	}

	time.Sleep(time.Second * 5)

	stopResult := removeContainer(dockerTask, createResult.ContainerId)
	if stopResult.Error != nil {
		fmt.Printf("%v", stopResult.Error)
		os.Exit(1)
	}
}

func createContainer() (*tasks.Docker, *tasks.DockerResult) {
	taskConfig := tasks.Config{
		Name:  "test-container-1",
		Image: "postgres:13",
		Env: []string{
			"POSTGRES_USER=cube",
			"POSTGRES_PASSWORD=secret",
		}}

	dockerClient, err := client.NewClientWithOpts(client.WithHost("unix:///Users/shamanskiy/.docker/run/docker.sock"))
	if err != nil {
		fmt.Printf("%v\n", err)
		return nil, nil
	}

	docker := tasks.Docker{
		Client: dockerClient,
		Config: taskConfig}

	result := docker.Run()
	if result.Error != nil {
		fmt.Printf("%v\n", result.Error)
		return nil, nil
	}

	log.Printf(
		"Container %s is running with config %+v\n", result.ContainerId, taskConfig)
	return &docker, &result
}

func removeContainer(docker *tasks.Docker, containerId string) *tasks.DockerResult {
	result := docker.Remove(containerId)
	if result.Error != nil {
		log.Printf("%v\n", result.Error)
		return nil
	}

	log.Printf(
		"Container %s has been stopped and removed\n", containerId)
	return &result
}
