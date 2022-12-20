package resweave_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/agilitree/resweave"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
		It("should fetch the index file directory correctly", func() {
			data, err := os.ReadFile(htmlDir + "index.html")
			Expect(err).ToNot(HaveOccurred())
			expContents := string(data)
			recorder := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, "localhost:8080/", nil)
			Expect(err).ToNot(HaveOccurred())

			htmlRes.Fetch(recorder, req)
			response := recorder.Result()
			defer response.Body.Close()
			Expect(response.StatusCode).To(Equal(http.StatusOK))
			respData, err := io.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(respData)).To(Equal(expContents))
		})
	})
})
