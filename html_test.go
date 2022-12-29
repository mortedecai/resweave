package resweave_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/mortedecai/resweave"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

const (
	htmlDir = "testing/html/"
)

var _ = Describe("Html", func() {
	var _ = Describe("Unnamed (root) resource", func() {
		var (
			htmlRes resweave.HTMLResource
			name    resweave.ResourceName
		)
		BeforeEach(func() {
			name = ""
			htmlRes = resweave.NewHTML(name, htmlDir)
		})
		It("should be possible to create an HTML resource for a local directory", func() {
			Expect(htmlRes).ToNot(BeNil())
		})
		It("should be possible to get the resource name", func() {
			Expect(htmlRes.Name()).To(Equal(name))
		})
		It("should be possible to check the base directory of the HTML resource", func() {
			Expect(htmlRes.BaseDir()).To(Equal(htmlDir))
		})
		It("should have a nil logger initially", func() {
			Expect(htmlRes.Logger()).To(BeNil())
		})
		It("should be possible to set a logger", func() {
			// It doesn't matter if recursive is true or false, HTMLResources cannot have child Resources.
			Expect(htmlRes.Logger()).To(BeNil())
			l, err := zap.NewProduction()
			Expect(err).ToNot(HaveOccurred())
			htmlRes.SetLogger(l.Sugar(), true)
			Expect(htmlRes.Logger()).ToNot(BeNil())
			htmlRes.SetLogger(nil, true)
			Expect(htmlRes.Logger()).To(BeNil())
		})
		It("should fetch the index file directory correctly", func() {
			data, err := os.ReadFile(htmlDir + "index.html")
			Expect(err).ToNot(HaveOccurred())
			expContents := string(data)
			recorder := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, "/", nil)
			Expect(err).ToNot(HaveOccurred())

			htmlRes.Fetch(recorder, req)
			response := recorder.Result()
			defer response.Body.Close()
			Expect(response.StatusCode).To(Equal(http.StatusOK))
			respData, err := io.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(respData)).To(Equal(expContents))
		})
		It("should return the correct fullPath() value", func() {
			Expect(htmlRes.FullPath()).To(Equal(resweave.ResourceName("/")))
			htmlRes = resweave.NewHTML("/one", htmlDir)
			Expect(htmlRes.FullPath()).To(Equal(resweave.ResourceName("/one/")))
		})
	})
	var _ = Describe("Named root resource", func() {
		var (
			htmlRes resweave.HTMLResource
			name    resweave.ResourceName
		)
		BeforeEach(func() {
			name = "name"
			htmlRes = resweave.NewHTML(name, htmlDir)
		})
		It("should be possible to create a named resource", func() {
			Expect(htmlRes.Name()).To(Equal(name))
		})
		It("should be possible to fetch the index of a named resource", func() {
			data, err := os.ReadFile(htmlDir + "index.html")
			Expect(err).ToNot(HaveOccurred())
			expContents := string(data)
			recorder := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/%s/", name), nil)
			Expect(err).ToNot(HaveOccurred())

			htmlRes.Fetch(recorder, req)
			response := recorder.Result()
			defer response.Body.Close()
			Expect(response.StatusCode).To(Equal(http.StatusOK))
			respData, err := io.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(respData)).To(Equal(expContents))
		})
		It("should return the correct fullPath() value", func() {
			expPath := resweave.ResourceName("/" + name + "/")
			Expect(htmlRes.FullPath()).To(Equal(expPath))
		})
	})
})
