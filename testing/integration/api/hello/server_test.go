package main_test

import (
	"fmt"
	"github.com/mortedecai/go-go-gadgets/env"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"io"
	"net/http"
)

var _ = Describe("Hello", func() {
	It("should be possible to receive \"Hello, World!\" from the hello LIST endpoint", func() {
		td := createTestData("Hello, World!\nRequest: '[a-z0-9]{8}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{12}'")
		td.RunTest()
	})
})

type TestData interface {
	RunTest()
}

func createTestData(exp string) TestData {
	host, _ := env.GetWithDefault("HOST_NAME", "localhost")
	return resweaveAPITestData{data: exp, host: host}
}

type resweaveAPITestData struct {
	data string
	host string
}

func (r resweaveAPITestData) RunTest() {
	r.runTest(fmt.Sprintf("http://%s:8080/hello", r.host))
}

func (r resweaveAPITestData) runTest(uri string) {
	response, err := http.Get(uri)
	Expect(err).ToNot(HaveOccurred())
	defer response.Body.Close()
	Expect(response.StatusCode).To(Equal(http.StatusOK))
	respData, err := io.ReadAll(response.Body)
	Expect(err).ToNot(HaveOccurred())
	Expect(string(respData)).To(MatchRegexp(r.data))
}
