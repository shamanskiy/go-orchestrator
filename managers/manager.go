package managers

import (
	"fmt"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
	"github.com/shamanskiy/go-orchestrator/tasks"
)

type Manager struct {
	Pending       queue.Queue
	TaskDb        map[string][]tasks.Task
	EventDb       map[string][]tasks.TaskEvent
	Workers       []string
	WorkerTaskMap map[string][]uuid.UUID
	TaskWorkerMap map[uuid.UUID]string
}

func (m *Manager) SelectWorker() {
	fmt.Println("I will select an appropriate worker")
}
func (m *Manager) UpdateTasks() {
	fmt.Println("I will update tasks")
}
func (m *Manager) SendWork() {
	fmt.Println("I will send work to workers")
}
