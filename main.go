package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/docker/docker/client"
	"github.com/google/uuid"
	"github.com/shamanskiy/go-orchestrator/common/queues"
	"github.com/shamanskiy/go-orchestrator/tasks"
	"github.com/shamanskiy/go-orchestrator/workers"
)

func main() {
	log.Println(uuid.New())

	worker := newWorker()
	go worker.ProcessTasksRequests(time.Second * 10)

	workerApi := newWorkerApi(worker)
	workerApi.Listen()
}

func newWorker() *workers.Worker {
	return &workers.Worker{
		TaskDb:           make(map[uuid.UUID]tasks.Task),
		TaskRequestQueue: queues.New[tasks.TaskRequest](),
		DockerClient:     newDockerClient(),
	}
}

func newWorkerApi(worker *workers.Worker) *workers.API {
	host := os.Getenv("CUBE_HOST")
	port, _ := strconv.Atoi(os.Getenv("CUBE_PORT"))
	return workers.NewAPI(host, port, worker)
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
