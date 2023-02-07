package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/agilitree/resweave"
	"github.com/mortedecai/go-go-gadgets/env"
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
	Description string     `json:"description"`
}

// TodoResource is an APIResource for handling TODOs
type TodoResource struct {
	resweave.APIResource
	todos []Todo
}

func createTodoResource(name resweave.ResourceName) (*TodoResource, error) {
	res := &TodoResource{
		resweave.NewAPI(name),
		make([]Todo, 0),
	}
	res.SetCreate(res.createTodo)
	res.SetList(res.listTodos)
	if err := res.SetID(resweave.NumericID); err != nil {
		return nil, err
	}
	res.SetFetch(res.fetchTodo)

	return res, nil
}

func (tr *TodoResource) fetchTodo(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	const curMethod = "fetchTodos"
	var id int
	var err error
	var bytes []byte = make([]byte, 0)
	if req.Method != http.MethodGet {
		tr.Infow(curMethod, "Bad Method", req.Method, "Accepted Method(s)", http.MethodGet)
		w.WriteHeader(http.StatusInternalServerError)
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
	for _, v := range tr.todos {
		tr.Infow(curMethod, "Fetching", id, "current", (*v.ID))
		if (*v.ID) == id {
			if bytes, err = json.Marshal(v); err != nil {
				tr.Infow(curMethod, "ID", id, "Marshall Error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			break
		}
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
	if dataBytes, err := io.ReadAll(req.Body); err != nil {
		tr.Infow(curMethod, "Data Read Error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		if err := json.Unmarshal(dataBytes, &todo); err != nil {
			tr.Infow(curMethod, "Unmarshall Error", err, "Incoming Data", string(dataBytes))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	if todo.ID == nil {
		id := (len(tr.todos) + 1)
		todo.ID = &id
	}
	if dataBytes, err := json.Marshal(todo); err != nil {
		tr.Infow(curMethod, "Marshall Return Error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		tr.todos = append(tr.todos, todo)
		w.WriteHeader(http.StatusCreated)
		if bw, err := w.Write(dataBytes); err != nil {
			tr.Infow(curMethod, "Response Write Error", err, "Bytes Written", bw)
		}
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
