package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	pkg "github.com/agilitree/resweave/testing/integration/api/todo"
	"github.com/mortedecai/go-go-gadgets/env"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Todos", Ordered, func() {
	var (
		host  string
		found bool
		todos []pkg.Todo
	)
	BeforeEach(func() {
		host, found = env.GetWithDefault("HOST_NAME", "localhost")
		if found {
			fmt.Println("Found hostname in environment")
		}
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
		var expID int = 1
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
		var expID int = 1
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

})
