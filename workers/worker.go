package workers

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/shamanskiy/go-orchestrator/queues"
	"github.com/shamanskiy/go-orchestrator/tasks"
)

type Worker struct {
	Name             string
	TaskRequestQueue *queues.Queue[tasks.TaskRequest]
	TaskDb           map[uuid.UUID]tasks.Task

	DockerClient *tasks.Docker
}

func (w *Worker) CollectStats() {
	fmt.Println("I will collect stats")
}

func (w *Worker) SubmitTaskRequest(request tasks.TaskRequest) {
	w.TaskRequestQueue.Enqueue(request)
}

func (w *Worker) ProcessTaskRequest() tasks.DockerResult {
	taskRequest, ok := w.TaskRequestQueue.Dequeue()
	if !ok {
		log.Println("no task requests to process")
		return tasks.DockerResult{}
	}

	switch taskRequest.RequiredState {
	case tasks.Scheduled:
		return w.startTask(taskRequest)
	case tasks.Completed:
		return w.stopTask(taskRequest.ID)
	default:
		err := fmt.Errorf("task transition not implemented: %v", taskRequest.RequiredState)
		return tasks.DockerResult{Error: err}
	}
}

func (w *Worker) startTask(request tasks.TaskRequest) tasks.DockerResult {
	task, ok := w.TaskDb[request.ID]
	if ok {
		err := fmt.Errorf("task %s already exists", request.ID)
		return tasks.DockerResult{Error: err}
	}

	task = request.Task()
	task.StartTime = time.Now().UTC()

	result := w.DockerClient.Run(task.Config())
	if result.Error != nil {
		task.State = tasks.Failed
		w.TaskDb[task.ID] = task
		return result
	}

	task.ContainerID = result.ContainerId
	task.State = tasks.Running
	w.TaskDb[task.ID] = task

	log.Printf("task %s started as container %s", task.ID, task.ContainerID)
	return result
}

func (w *Worker) stopTask(taskId uuid.UUID) tasks.DockerResult {
	task, ok := w.TaskDb[taskId]
	if !ok {
		err := fmt.Errorf("task %s not found", taskId)
		return tasks.DockerResult{Error: err}
	}

	if task.State != tasks.Running {
		err := fmt.Errorf("task %s is not running, current state: %v", task.ID, task.State)
		return tasks.DockerResult{Error: err}
	}

	result := w.DockerClient.Remove(task.ContainerID)
	if result.Error != nil {
		return result
	}

	task.FinishTime = time.Now().UTC()
	task.State = tasks.Completed
	w.TaskDb[task.ID] = task
	log.Printf("task %s completed at %s", task.ID, task.FinishTime)

	return result
}
