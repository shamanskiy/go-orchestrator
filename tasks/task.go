package tasks

import (
	"time"

	"github.com/docker/docker/api/types/container"
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

	Runtime TaskRuntime
}

type TaskRuntime struct {
	ContainerID string
	Port        int
	State       State
	StartTime   time.Time
	FinishTime  time.Time
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
}
