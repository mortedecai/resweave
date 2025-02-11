package main_test

import (
	"fmt"
	"io"
	"net/http"

	"github.com/mortedecai/go-go-gadgets/env"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Hello", func() {
	It("should be possible to receive \"Hello, World!\" from the hello LIST endpoint", func() {
		td := createTestData("Hello, World!\nRequest: '[a-z0-9]{8}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{12}'")
		td.RunTest("http://%s:8080/hello")
		td = createTestData("Bonjour, Toute le monde!\nRequest: '[a-z0-9]{8}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{12}'")
		td.RunTest("http://%s:8080/hello/translate")
	})
})

type TestData interface {
	RunTest(string)
}

func createTestData(exp string) TestData {
	host, _ := env.GetWithDefault("HOST_NAME", "localhost")
	return resweaveAPITestData{data: exp, host: host}
}

type resweaveAPITestData struct {
	data string
	host string
}

func (r resweaveAPITestData) RunTest(formatableURI string) {
	r.runTest(fmt.Sprintf(formatableURI, r.host))
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
