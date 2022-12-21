package resweave

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Host", func() {
	const (
		caHostName = HostName("daniel-taylor.ca")
	)
	var (
		caHost Host
	)
	BeforeEach(func() {
		caHost = newHost(caHostName)
	})
	Describe("Initialization", func() {
		It("should initialize a non-null Host implementation", func() {
			Expect(caHost).ToNot(BeNil())
		})
		It("should have the provided name", func() {
			Expect(caHost.Name()).To(Equal(caHostName))
		})
		It("should have an empty resource map", func() {
			Expect(caHost.TopLevelResourceCount()).To(BeZero())
		})
	})
	Describe("Usage", func() {
		const (
			htmlDir = "testing/html/"
		)
		It("should be possible to add an unnamed resource", func() {
			Expect(caHost.AddResource(NewHTML("", htmlDir))).ToNot(HaveOccurred())
		})
		It("should increase the resource count when adding an unnamed resource", func() {
			Expect(caHost.AddResource(NewHTML("", htmlDir))).ToNot(HaveOccurred())
			Expect(caHost.TopLevelResourceCount()).To(Equal(1))
		})
		It("should be possible to retrieve the resource after adding an unnamed resource", func() {
			htmlRes := NewHTML("", htmlDir)
			Expect(caHost.AddResource(htmlRes)).ToNot(HaveOccurred())
			Expect(caHost.TopLevelResourceCount()).To(Equal(1))
			res, found := caHost.GetResource("")
			Expect(found).To(BeTrue())
			Expect(res).To(Equal(htmlRes))
		})
		It("should return an error if two unnamed resources are added", func() {
			htmlRes := NewHTML("", htmlDir)
			Expect(caHost.AddResource(htmlRes)).ToNot(HaveOccurred())
			Expect(caHost.TopLevelResourceCount()).To(Equal(1))
			Expect(caHost.AddResource(htmlRes)).To(HaveOccurred())
			Expect(caHost.AddResource(htmlRes)).To(Equal(fmt.Errorf(FmtResourceAlreadyExists, caHost.Name())))
			Expect(caHost.TopLevelResourceCount()).To(Equal(1))
		})
		It("should serve an unnamed root resource correctly", func() {
			Expect(caHost.AddResource(NewHTML("", htmlDir))).ToNot(HaveOccurred())
			data, err := os.ReadFile(htmlDir + "index.html")
			Expect(err).ToNot(HaveOccurred())
			expContents := string(data)
			recorder := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, "localhost:8080/", nil)
			Expect(err).ToNot(HaveOccurred())

			caHost.Serve(recorder, req)
			response := recorder.Result()
			defer response.Body.Close()
			Expect(response.StatusCode).To(Equal(http.StatusOK))
			respData, err := io.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(respData)).To(Equal(expContents))

		})
		It("should should return a 404 if no resources were added", func() {
			recorder := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, "localhost:8080/", nil)
			Expect(err).ToNot(HaveOccurred())

			caHost.Serve(recorder, req)
			response := recorder.Result()
			defer response.Body.Close()
			Expect(response.StatusCode).To(Equal(http.StatusNotFound))
		})
	})
})
