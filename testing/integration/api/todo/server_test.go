package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mortedecai/go-go-gadgets/env"
	pkg "github.com/mortedecai/resweave/testing/integration/api/todo"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Todos", Ordered, func() {
	var (
		host  string
		todos []pkg.Todo
	)
	BeforeEach(func() {
		host, _ = env.GetWithDefault("HOST_NAME", "localhost")
		todos = make([]pkg.Todo, 0)
	})
	It("should be possible to retrieve an empty list of todos", func() {
		uri := fmt.Sprintf("http://%s:8080/todos", host)
		response, err := http.Get(uri)
		Expect(err).ToNot(HaveOccurred())
		defer response.Body.Close()

		Expect(response.StatusCode).To(Equal(http.StatusOK))
		respData, err := io.ReadAll(response.Body)
		Expect(err).ToNot(HaveOccurred())

		err = json.Unmarshal(respData, &todos)
		Expect(err).ToNot(HaveOccurred())
		Expect(len(todos)).To(BeZero())
	})
	It("should be possible to create a todo", func() {
		const desc = "Simple Todo"
		var expID int = 0
		var postBytes []byte
		var err error
		var recTodo pkg.Todo
		expTodo := pkg.Todo{ID: &expID, Description: desc}

		uri := fmt.Sprintf("http://%s:8080/todos", host)
		todo := pkg.Todo{Description: desc}
		postBytes, err = json.Marshal(&todo)
		Expect(err).ToNot(HaveOccurred())

		buffer := bytes.NewBuffer(postBytes)
		response, err := http.Post(uri, "application/json", buffer)
		Expect(err).ToNot(HaveOccurred())
		defer response.Body.Close()

		Expect(response.StatusCode).To(Equal(http.StatusCreated))
		respData, err := io.ReadAll(response.Body)
		Expect(err).ToNot(HaveOccurred())

		err = json.Unmarshal(respData, &recTodo)
		Expect(err).ToNot(HaveOccurred())
		Expect(recTodo).To(Equal(expTodo))
	})
	It("should be possible to retrieve a non-empty list of todos", func() {
		const desc = "Simple Todo"
		var expID int = 0
		var err error
		expTodo := pkg.Todo{ID: &expID, Description: desc}

		uri := fmt.Sprintf("http://%s:8080/todos", host)
		response, err := http.Get(uri)
		Expect(err).ToNot(HaveOccurred())
		defer response.Body.Close()

		Expect(response.StatusCode).To(Equal(http.StatusOK))
		respData, err := io.ReadAll(response.Body)
		Expect(err).ToNot(HaveOccurred())

		err = json.Unmarshal(respData, &todos)
		Expect(err).ToNot(HaveOccurred())
		Expect(len(todos)).To(Equal(1))
		Expect(todos[0]).To(Equal(expTodo))
	})
	It("should be possible to retrieve a single todo from a list", func() {
		const desc = "Simple Todo"
		var expID int = 0
		var err error
		expTodo := pkg.Todo{ID: &expID, Description: desc}

		uri := fmt.Sprintf("http://%s:8080/todos", host)
		response, err := http.Get(uri)
		Expect(err).ToNot(HaveOccurred())
		defer response.Body.Close()

		Expect(response.StatusCode).To(Equal(http.StatusOK))
		respData, err := io.ReadAll(response.Body)
		Expect(err).ToNot(HaveOccurred())

		err = json.Unmarshal(respData, &todos)
		Expect(err).ToNot(HaveOccurred())
		Expect(len(todos)).To(Equal(1))
		Expect(todos[0]).To(Equal(expTodo))

		fetchUri := fmt.Sprintf("http://%s:8080/todos/%d", host, (*todos[0].ID))
		fetchResponse, fetchErr := http.Get(fetchUri)
		Expect(fetchErr).ToNot(HaveOccurred())
		defer fetchResponse.Body.Close()

		//Expect(fetchResponse.StatusCode).To(Equal(http.StatusTeapot))
		Expect(fetchResponse.StatusCode).To(Equal(http.StatusOK))
		fetchRespData, err := io.ReadAll(fetchResponse.Body)
		Expect(err).ToNot(HaveOccurred())
		var fetchedTodo pkg.Todo
		err = json.Unmarshal(fetchRespData, &fetchedTodo)
		Expect(err).ToNot(HaveOccurred())
		Expect(fetchedTodo).To(Equal(todos[0]))
	})
	It("should be possible to PUT a single todo from a list", func() {
		const newDesc = "This used to be a Simple Todo"
		var expID int = 0
		var err error
		var todo pkg.Todo
		var putData []byte

		uri := fmt.Sprintf("http://%s:8080/todos/", host)
		response, err := http.Get(uri)
		Expect(err).ToNot(HaveOccurred())
		defer response.Body.Close()

		Expect(response.StatusCode).To(Equal(http.StatusOK))
		respData, err := io.ReadAll(response.Body)
		Expect(err).ToNot(HaveOccurred())

		err = json.Unmarshal(respData, &todos)
		Expect(err).ToNot(HaveOccurred())
		Expect(len(todos)).To(Equal(1))

		todo = todos[0]
		todo.Description = newDesc

		putData, err = json.Marshal(todo)
		Expect(err).ToNot(HaveOccurred())
		buff := bytes.NewBuffer(putData)

		putUri := fmt.Sprintf("http://%s:8080/todos/%d", host, (*todo.ID))
		req, err := http.NewRequest(http.MethodPut, putUri, buff)
		Expect(err).ToNot(HaveOccurred())
		putResponse, putErr := http.DefaultClient.Do(req)
		Expect(putErr).ToNot(HaveOccurred())
		defer putResponse.Body.Close()

		Expect(putResponse.StatusCode).To(Equal(http.StatusOK))
		putData, err = io.ReadAll(putResponse.Body)
		Expect(err).ToNot(HaveOccurred())
		var putedTodo pkg.Todo
		err = json.Unmarshal(putData, &putedTodo)
		Expect(err).ToNot(HaveOccurred())
		Expect(putedTodo.Description).To(Equal(newDesc))
		Expect(*putedTodo.ID).To(Equal(expID))

		fetchUri := fmt.Sprintf("http://%s:8080/todos/%d", host, (*todos[0].ID))
		fetchResponse, fetchErr := http.Get(fetchUri)
		Expect(fetchErr).ToNot(HaveOccurred())
		defer fetchResponse.Body.Close()

		Expect(fetchResponse.StatusCode).To(Equal(http.StatusOK))
		fetchRespData, err := io.ReadAll(fetchResponse.Body)
		Expect(err).ToNot(HaveOccurred())
		var fetchedTodo pkg.Todo
		err = json.Unmarshal(fetchRespData, &fetchedTodo)
		Expect(err).ToNot(HaveOccurred())
		Expect(fetchedTodo).To(Equal(putedTodo))
	})

	It("should be possible to PATCH a single todo from a list", func() {
		const expDesc = "This used to be a Simple Todo"
		var expID int = 0
		var err error
		var todo pkg.Todo = pkg.Todo{ID: new(int)}
		var patchData []byte

		*todo.ID = 0
		todo.Completed = true

		patchData, err = json.Marshal(todo)
		Expect(err).ToNot(HaveOccurred())
		buff := bytes.NewBuffer(patchData)

		patchUri := fmt.Sprintf("http://%s:8080/todos/%d", host, (*todo.ID))
		req, err := http.NewRequest(http.MethodPatch, patchUri, buff)
		Expect(err).ToNot(HaveOccurred())
		patchResponse, patchErr := http.DefaultClient.Do(req)
		Expect(patchErr).ToNot(HaveOccurred())
		defer patchResponse.Body.Close()

		Expect(patchResponse.StatusCode).To(Equal(http.StatusOK))
		patchData, err = io.ReadAll(patchResponse.Body)
		Expect(err).ToNot(HaveOccurred())
		var patchedTodo pkg.Todo
		err = json.Unmarshal(patchData, &patchedTodo)
		Expect(err).ToNot(HaveOccurred())
		Expect(patchedTodo.Description).To(Equal(expDesc))
		Expect(*patchedTodo.ID).To(Equal(expID))

		fetchUri := fmt.Sprintf("http://%s:8080/todos/%d", host, *todo.ID)
		fetchResponse, fetchErr := http.Get(fetchUri)
		Expect(fetchErr).ToNot(HaveOccurred())
		defer fetchResponse.Body.Close()

		Expect(fetchResponse.StatusCode).To(Equal(http.StatusOK))
		fetchRespData, err := io.ReadAll(fetchResponse.Body)
		Expect(err).ToNot(HaveOccurred())
		var fetchedTodo pkg.Todo
		err = json.Unmarshal(fetchRespData, &fetchedTodo)
		Expect(err).ToNot(HaveOccurred())
		Expect(fetchedTodo).To(Equal(patchedTodo))
		Expect(fetchedTodo).ToNot(Equal(todo))
		Expect(*fetchedTodo.ID).To(Equal(*todo.ID))
		Expect(fetchedTodo.Completed).To(Equal(todo.Completed))
		Expect(fetchedTodo.Description).To(Equal(expDesc))

	})

	It("should be possible to delete a single todo from a list", func() {
		const desc = "This used to be a Simple Todo"
		var expID int = 0
		var err error
		expTodo := pkg.Todo{ID: &expID, Description: desc, Completed: true}
		uri := fmt.Sprintf("http://%s:8080/todos", host)
		response, err := http.Get(uri)
		Expect(err).ToNot(HaveOccurred())
		defer response.Body.Close()

		Expect(response.StatusCode).To(Equal(http.StatusOK))
		respData, err := io.ReadAll(response.Body)
		Expect(err).ToNot(HaveOccurred())

		err = json.Unmarshal(respData, &todos)
		Expect(err).ToNot(HaveOccurred())
		Expect(len(todos)).To(Equal(1))
		Expect(todos[0]).To(Equal(expTodo))

		deleteUri := fmt.Sprintf("http://%s:8080/todos/%d", host, (*todos[0].ID))
		req, err := http.NewRequest(http.MethodDelete, deleteUri, nil)
		Expect(err).ToNot(HaveOccurred())
		deleteResponse, deleteErr := http.DefaultClient.Do(req)
		Expect(deleteErr).ToNot(HaveOccurred())
		defer deleteResponse.Body.Close()

		Expect(deleteResponse.StatusCode).To(Equal(http.StatusNoContent))

		fetchUri := fmt.Sprintf("http://%s:8080/todos/%d", host, (*todos[0].ID))
		fetchResponse, fetchErr := http.Get(fetchUri)
		Expect(fetchErr).ToNot(HaveOccurred())
		defer fetchResponse.Body.Close()

		Expect(fetchResponse.StatusCode).To(Equal(http.StatusNotFound))
	})
})
