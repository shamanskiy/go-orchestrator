package workers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/shamanskiy/go-orchestrator/common"
	"github.com/shamanskiy/go-orchestrator/tasks"
)

type API struct {
	Worker *Worker
	Router *http.ServeMux

	Host string
	Port int
}

func NewAPI(host string, port int, worker *Worker) *API {
	api := &API{
		Host:   host,
		Port:   port,
		Worker: worker,
		Router: http.NewServeMux(),
	}

	api.Router.HandleFunc("GET /tasks", api.GetTasksHandler)
	api.Router.HandleFunc("POST /tasks", api.PostTaskHandler)
	api.Router.HandleFunc("DELETE /tasks/{taskId}", api.DeleteTaskHandler)

	return api
}

func (a *API) Listen() {
	addr := fmt.Sprintf("%s:%d", a.Host, a.Port)
	fmt.Printf("Starting server on %s\n", addr)
	http.ListenAndServe(addr, a.Router)
}

func (a *API) GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(a.Worker.Tasks())
}

type PostTaskRequest struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Image string    `json:"image"`
}

func (r PostTaskRequest) toTaskRequest() tasks.TaskRequest {
	return tasks.TaskRequest{
		ID:            r.ID,
		Name:          r.Name,
		Image:         r.Image,
		RequiredState: tasks.Scheduled,
	}
}

func (a *API) PostTaskHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	postTaskRequest := PostTaskRequest{}
	err := decoder.Decode(&postTaskRequest)
	if err != nil {
		msg := fmt.Sprintf("Error unmarshalling body: %v\n", err)
		apiError := common.ApiError{
			Code:    400,
			Message: msg,
		}

		log.Println(msg)
		w.WriteHeader(apiError.Code)
		json.NewEncoder(w).Encode(apiError)
		return
	}

	a.Worker.SubmitTaskRequest(postTaskRequest.toTaskRequest())
	log.Printf("submitted task request: %v\n", postTaskRequest.toTaskRequest())
	w.WriteHeader(204)
}

func (a *API) DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	taskId := r.PathValue("taskId")
	if taskId == "" {
		msg := "taskId is required"
		apiError := common.ApiError{
			Code:    400,
			Message: msg,
		}

		log.Println(msg)
		w.WriteHeader(apiError.Code)
		json.NewEncoder(w).Encode(apiError)
		return
	}

	id, err := uuid.Parse(taskId)
	if err != nil {
		msg := fmt.Sprintf("Error parsing taskId: %v\n", err)
		apiError := common.ApiError{
			Code:    400,
			Message: msg,
		}

		log.Println(msg)
		w.WriteHeader(apiError.Code)
		json.NewEncoder(w).Encode(apiError)
		return
	}

	request := tasks.TaskRequest{
		ID:            id,
		RequiredState: tasks.Completed,
	}
	a.Worker.SubmitTaskRequest(request)
	log.Printf("submitted task request: %v\n", request)
	w.WriteHeader(204)
}
