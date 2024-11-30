package main_test

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Hello", func() {
	It("should be possible to receive the hello world index page", func() {
		td := createTestData("index.html")
		td.RunTest()
		td.RunIndexTest()
	})
	It("should be possible to receive the hello world index page", func() {
		td := createTestData("goodbye.html")
		td.RunTest()
	})
})

type TestData interface {
	RunTest()
	RunIndexTest()
}

func createTestData(filename string) TestData {
	if host, found := os.LookupEnv("HOST_NAME"); found {
		return resweaveHtmlTestData{page: filename, host: host}
	}
	return resweaveHtmlTestData{page: filename, host: "localhost"}
}

type resweaveHtmlTestData struct {
	page string
	host string
}

func (r resweaveHtmlTestData) RunIndexTest() {
	if !strings.HasPrefix(r.page, "index.") {
		Skip("Not an index page")
	}
	r.runTest("http://" + r.host + ":8080/")

}

func (r resweaveHtmlTestData) RunTest() {
	r.runTest(fmt.Sprintf("http://"+r.host+":8080/%s", r.page))
}

func (r resweaveHtmlTestData) runTest(uri string) {
	data, err := os.ReadFile(fmt.Sprintf("html/%s", r.page))
	Expect(err).ToNot(HaveOccurred())
	expContents := string(data)

	response, err := http.Get(uri)
	Expect(err).ToNot(HaveOccurred())
	defer response.Body.Close()
	Expect(response.StatusCode).To(Equal(http.StatusOK))
	respData, err := io.ReadAll(response.Body)
	Expect(err).ToNot(HaveOccurred())
	Expect(string(respData)).To(Equal(expContents))
}
