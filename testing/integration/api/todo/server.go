package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/mortedecai/go-go-gadgets/env"
	"github.com/mortedecai/resweave"
	"go.uber.org/zap"
)

const (
	varPort     = "RESWEAVE_PORT"
	defaultPort = 8080

	logStatus    = "Status"
	logCompleted = "Completed"
	logStarting  = "Starting"
)

// Todo struct holds the information for a Todo item.
type Todo struct {
	ID          *int       `json:"id,omitempty"`
	Due         *time.Time `json:"due,omitempty"`
	Completed   bool       `json:"completed,omitempty"`
	Description string     `json:"description,omitempty"`
}

// TodoResource is an APIResource for handling TODOs
type TodoResource struct {
	resweave.APIResource
	todos  []Todo
	nextID int
	mtx    sync.Mutex
}

func createTodoResource(name resweave.ResourceName) (*TodoResource, error) {
	res := &TodoResource{
		APIResource: resweave.NewAPI(name),
		todos:       make([]Todo, 0),
	}
	res.SetCreate(res.createTodo)
	res.SetList(res.listTodos)
	if err := res.SetID(resweave.NumericID); err != nil {
		return nil, err
	}
	res.SetFetch(res.fetchTodo)
	res.SetDelete(res.deleteTodo)
	res.SetUpdate(res.updateTodo)

	return res, nil
}

func (tr *TodoResource) handlePut(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	const curMethod = "handlePut"
	var id int
	var err error
	if req.Method != http.MethodPut && req.Method != http.MethodPatch {
		tr.Infow(curMethod, "Bad Method", req.Method, "Accepted Method(s)", http.MethodPut)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	v := ctx.Value(resweave.Key(fmt.Sprintf("id_%s", tr.Name().String())))
	var val string
	var ok bool
	if val, ok = v.(string); !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if id, err = strconv.Atoi(val); err != nil {
		tr.Infow(curMethod, "Bad ID", val, "Error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var updateData Todo
	dataBytes, err := io.ReadAll(req.Body)
	if err != nil {
		tr.Infow(curMethod, "Data Read Error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := json.Unmarshal(dataBytes, &updateData); err != nil {
		tr.Infow(curMethod, "Unmarshall Error", err, "Incoming Data", string(dataBytes))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tr.mtx.Lock()
	defer tr.mtx.Unlock()

	var respBytes []byte
	for i, v := range tr.todos {
		if (*v.ID) == id {
			tr.Infow(curMethod, "Updating", id, "current", v, "update", updateData)
			// Ordering in the array doesn't matter; lookup by ID value
			tr.todos[i].Due = updateData.Due
			tr.todos[i].Completed = updateData.Completed
			tr.todos[i].Description = updateData.Description
			tr.Infow(curMethod, "Updated", id, "current", tr.todos[i], "update", updateData)
			respBytes, err = json.Marshal(tr.todos[i])
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			break
		}
	}
	w.WriteHeader(http.StatusOK)
	bw, err := w.Write(respBytes)
	if err != nil {
		tr.Infow(curMethod, "Updating", id, "Write Error", err, "Bytes Written", bw)
	}
}

func (tr *TodoResource) handlePatch(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	const curMethod = "handlePatch"
	var id int
	var err error
	if req.Method != http.MethodPut && req.Method != http.MethodPatch {
		tr.Infow(curMethod, "Bad Method", req.Method, "Accepted Method(s)", http.MethodPut)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	v := ctx.Value(resweave.Key(fmt.Sprintf("id_%s", tr.Name().String())))
	var val string
	var ok bool
	if val, ok = v.(string); !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if id, err = strconv.Atoi(val); err != nil {
		tr.Infow(curMethod, "Bad ID", val, "Error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var updateData Todo
	dataBytes, err := io.ReadAll(req.Body)
	if err != nil {
		tr.Infow(curMethod, "Data Read Error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := json.Unmarshal(dataBytes, &updateData); err != nil {
		tr.Infow(curMethod, "Unmarshall Error", err, "Incoming Data", string(dataBytes))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tr.mtx.Lock()
	defer tr.mtx.Unlock()

	var respBytes []byte
	for i, v := range tr.todos {
		if (*v.ID) == id {
			tr.Infow(curMethod, "Updating", id, "current", v, "update", updateData)
			// Ordering in the array doesn't matter; lookup by ID value
			var emptyTime time.Time
			if updateData.Due != nil && !updateData.Due.Equal(emptyTime) {
				tr.todos[i].Due = updateData.Due
			}
			tr.todos[i].Completed = updateData.Completed
			if len(updateData.Description) > 0 {
				tr.todos[i].Description = updateData.Description
			}
			tr.Infow(curMethod, "Updated", id, "current", tr.todos[i], "update", updateData)
			respBytes, err = json.Marshal(tr.todos[i])
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			break
		}
	}
	w.WriteHeader(http.StatusOK)
	bw, err := w.Write(respBytes)
	if err != nil {
		tr.Infow(curMethod, "Updating", id, "Write Error", err, "Bytes Written", bw)
	}
}

func (tr *TodoResource) updateTodo(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPut:
		tr.handlePut(ctx, w, req)
	case http.MethodPatch:
		tr.handlePatch(ctx, w, req)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (tr *TodoResource) deleteTodo(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	const curMethod = "deleteTodo"
	var id int
	var err error
	if req.Method != http.MethodDelete {
		tr.Infow(curMethod, "Bad Method", req.Method, "Accepted Method(s)", http.MethodDelete)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	v := ctx.Value(resweave.Key(fmt.Sprintf("id_%s", tr.Name().String())))
	var val string
	var ok bool
	if val, ok = v.(string); !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if id, err = strconv.Atoi(val); err != nil {
		tr.Infow(curMethod, "Bad ID", val, "Error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	tr.mtx.Lock()
	defer tr.mtx.Unlock()
	for i, v := range tr.todos {
		tr.Infow(curMethod, "Deleting", id, "current", (*v.ID))
		if (*v.ID) == id {
			// Ordering in the array doesn't matter; lookup by ID value
			tr.todos[i] = tr.todos[len(tr.todos)-1]
			tr.todos = tr.todos[:len(tr.todos)-1]
			break
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

func (tr *TodoResource) fetchTodo(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	const curMethod = "fetchTodos"
	var id int
	var err error
	var bytes []byte = make([]byte, 0)
	found := false
	if req.Method != http.MethodGet {
		tr.Infow(curMethod, "Bad Method", req.Method, "Accepted Method(s)", http.MethodGet)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	v := ctx.Value(resweave.Key(fmt.Sprintf("id_%s", tr.Name().String())))
	var val string
	var ok bool
	if val, ok = v.(string); !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if id, err = strconv.Atoi(val); err != nil {
		tr.Infow(curMethod, "Bad ID", val, "Error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	tr.mtx.Lock()
	defer tr.mtx.Unlock()
	for _, v := range tr.todos {
		tr.Infow(curMethod, "Fetching", id, "current", (*v.ID))
		if (*v.ID) == id {
			if bytes, err = json.Marshal(v); err != nil {
				tr.Infow(curMethod, "ID", id, "Marshall Error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			found = true
			break
		}
	}
	if !found {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	if bw, err := w.Write(bytes); err != nil {
		tr.Debugw(curMethod, "Error Writing", err, "Bytes Written", bw)
	}
}

func (tr *TodoResource) listTodos(_ context.Context, w http.ResponseWriter, req *http.Request) {
	const curMethod = "listTodos"
	var dataBytes []byte
	var err error
	tr.Infow(curMethod, logStatus, logStarting)
	if req.Method != http.MethodGet {
		tr.Infow(curMethod, "Bad Method", req.Method, "Accepted Method(s)", http.MethodGet)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tr.mtx.Lock()
	defer tr.mtx.Unlock()
	if dataBytes, err = json.Marshal(tr.todos); err != nil {
		tr.Infow(curMethod, "Unable To Marchal", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if bw, err := w.Write(dataBytes); err != nil {
		tr.Infow(curMethod, "Response Write Error", err, "Bytes Written", bw)
	}
	tr.Infow(curMethod, logStatus, logCompleted)
}

func (tr *TodoResource) createTodo(_ context.Context, w http.ResponseWriter, req *http.Request) {
	const curMethod = "createTodo"
	tr.Infow(curMethod, logStatus, logStarting)
	if req.Method != http.MethodPost {
		tr.Infow(curMethod, "Bad Method", req.Method, "Accepted Method(s)", http.MethodPost)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tr.Infow(curMethod, "Supported Method", req.Method)
	var todo Todo
	dataBytes, err := io.ReadAll(req.Body)
	if err != nil {
		tr.Infow(curMethod, "Data Read Error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := json.Unmarshal(dataBytes, &todo); err != nil {
		tr.Infow(curMethod, "Unmarshall Error", err, "Incoming Data", string(dataBytes))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	tr.mtx.Lock()
	defer tr.mtx.Unlock()
	if todo.ID == nil {
		id := tr.nextID
		tr.nextID++
		todo.ID = &id
	}
	dataBytes, err = json.Marshal(todo)
	if err != nil {
		tr.Infow(curMethod, "Marshall Return Error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tr.todos = append(tr.todos, todo)
	w.WriteHeader(http.StatusCreated)
	if bw, err := w.Write(dataBytes); err != nil {
		tr.Infow(curMethod, "Response Write Error", err, "Bytes Written", bw)
	}
	tr.Infow(curMethod, logStatus, logCompleted)
}

func main() {
	fmt.Println("Running Server for API Integration Test: TODO")

	var logger *zap.SugaredLogger
	var todoResource *TodoResource
	var err error
	port, _ := env.GetWithDefaultInt(varPort, defaultPort)
	server := resweave.NewServer(port)

	if l, err := zap.NewDevelopment(); err != nil {
		fmt.Println("******** COULD NOT CREATE A LOGGER!!!!!!! ************")
	} else {
		logger = l.Sugar()
		server.SetLogger(logger, true)
	}

	if todoResource, err = createTodoResource("todos"); err != nil {
		logger.Errorw("main", "createTodoResource", err)
		return
	}
	todoResource.SetLogger(logger, false)

	if err := server.AddResource(todoResource); err == nil {
		fmt.Println(server.Run())
	} else {
		fmt.Println(err.Error())
	}
}
