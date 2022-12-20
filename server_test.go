package resweave

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Server", func() {
	Describe("Initialization", func() {
		var (
			s Server
		)
		const (
			port    = 8080
			htmlDir = "testing/html/"
		)
		BeforeEach(func() {
			s = NewServer(port).(*server)
		})
		It("should be possible to create a new server", func() {
			Expect(s).ToNot(BeNil())
			Expect(s.Port()).To(Equal(port))
		})
		It("should be possible to create a new http.Server with the appropriate timeouts", func() {
			srv := s.(*server).createHTTPServer()
			Expect(srv).ToNot(BeNil())
			Expect(srv.Addr).To(Equal(fmt.Sprintf(":%d", port)))
			Expect(srv.ReadHeaderTimeout).To(Equal(3 * time.Second))
		})
		It("should not be possible to add a nil Resource", func() {
			Expect(s.AddResource(nil)).To(HaveOccurred())
			Expect(s.AddResource(nil).Error()).To(Equal("cannot add a nil resource"))
		})
		It("should be possible to add an unnamed resource", func() {
			Expect(s.AddResource(NewHTML("", htmlDir))).ToNot(HaveOccurred())
			v, s := s.GetResource("")
			Expect(v).ToNot(BeNil())
			Expect(v.Name()).To(BeEquivalentTo(""))
			Expect(s).To(BeTrue())
		})
		It("should serve an unnamed root resource correctly", func() {
			Expect(s.AddResource(NewHTML("", htmlDir))).ToNot(HaveOccurred())
			data, err := os.ReadFile(htmlDir + "index.html")
			Expect(err).ToNot(HaveOccurred())
			expContents := string(data)
			recorder := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, "localhost:8080/", nil)
			Expect(err).ToNot(HaveOccurred())

			s.(*server).serve(recorder, req)
			response := recorder.Result()
			defer response.Body.Close()
			Expect(response.StatusCode).To(Equal(http.StatusOK))
			respData, err := io.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(respData)).To(Equal(expContents))

		})
	})
})
