package main

import (
	"log"
	"time"

	"github.com/docker/docker/client"
	"github.com/google/uuid"
	"github.com/shamanskiy/go-orchestrator/queues"
	"github.com/shamanskiy/go-orchestrator/tasks"
	"github.com/shamanskiy/go-orchestrator/workers"
)

func main() {
	worker := workers.Worker{
		TaskDb:           make(map[uuid.UUID]tasks.Task),
		TaskRequestQueue: queues.New[tasks.TaskRequest](),
		DockerClient:     newDockerClient(),
	}

	taskRequest := tasks.TaskRequest{
		ID:            uuid.New(),
		Name:          "test-task-1",
		Image:         "strm/helloworld-http",
		RequiredState: tasks.Scheduled,
	}
	worker.SubmitTaskRequest(taskRequest)

	processResult := worker.ProcessTaskRequest()
	if processResult.Error != nil {
		panic(processResult.Error)
	}

	sleepTime := time.Second * 3
	log.Printf("sleeping for %+v seconds\n", sleepTime)
	time.Sleep(sleepTime)

	taskRequest.RequiredState = tasks.Completed
	worker.SubmitTaskRequest(taskRequest)

	processResult = worker.ProcessTaskRequest()
	if processResult.Error != nil {
		panic(processResult.Error)
	}
}

func newDockerClient() *tasks.Docker {
	dockerClient, err := client.NewClientWithOpts(client.WithHost("unix:///Users/shamanskiy/.docker/run/docker.sock"))
	if err != nil {
		panic(err)
	}
	return &tasks.Docker{
		Client: dockerClient,
	}
}
