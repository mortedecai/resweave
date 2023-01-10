package resweave_test

import (
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/mortedecai/resweave"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

var _ = Describe("Api", func() {
	var _ = Describe("Initialization", func() {
		var (
			res resweave.APIResource
		)
		BeforeEach(func() {
			res = resweave.NewAPI("")
		})
		It("should create a new APIResource", func() {
			Expect(res).ToNot(BeNil())
		})
		It("should have a nil logger initially", func() {
			Expect(res.Logger()).To(BeNil())
		})
		It("should be possible to retrieve the name of the resource", func() {
			Expect(res).ToNot(BeNil())
			Expect(res.Name()).To(Equal(resweave.ResourceName("")))
			name := resweave.ResourceName("one")
			res = resweave.NewAPI(name)
			Expect(res).ToNot(BeNil())
			Expect(res.Name()).To(Equal(name))
		})
	})
	var _ = Describe("Logger handling", func() {
		var (
			res resweave.APIResource
		)
		BeforeEach(func() {
			res = resweave.NewAPI("")
		})
		It("should be possible to set a non-recursive logger on a resource with no sub-resources", func() {
			Expect(res.Logger()).To(BeNil())
			l, err := zap.NewProduction()
			Expect(err).ToNot(HaveOccurred())
			res.SetLogger(l.Sugar(), false)
			Expect(res.Logger()).ToNot(BeNil())
			res.SetLogger(nil, false)
			Expect(res.Logger()).To(BeNil())
		})
		It("should be possible to set a recursive logger on a resource with no sub-resources", func() {
			Expect(res.Logger()).To(BeNil())
			l, err := zap.NewProduction()
			Expect(err).ToNot(HaveOccurred())
			res.SetLogger(l.Sugar(), true)
			Expect(res.Logger()).ToNot(BeNil())
			res.SetLogger(nil, true)
			Expect(res.Logger()).To(BeNil())
		})
	})
	var _ = Describe("Request Handling", func() {
		var (
			res resweave.APIResource
		)
		BeforeEach(func() {
			res = resweave.NewAPI("")
		})
		It("should return a 405 if the List function has not been supplied", func() {
			recorder := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, "/", nil)
			Expect(err).ToNot(HaveOccurred())

			res.List(recorder, req)
			response := recorder.Result()
			defer response.Body.Close()
			Expect(response.StatusCode).To(Equal(http.StatusMethodNotAllowed))
		})
		It("should allow a List function to be called", func() {
			var s *zap.SugaredLogger
			if l, err := zap.NewDevelopment(); err == nil {
				s = l.Sugar()
			} else {
				Expect(err).ToNot(HaveOccurred())
			}

			res.SetList(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				respBytes := []byte("Hello, World!")
				if bw, err := w.Write(respBytes); err != nil {
					s.Infow("List", "WriteError", err, "BytesWritten", bw)
				} else {
					s.Debugw("List", "BytesWritten", bw)
				}
			})
			expContents := "Hello, World!"
			recorder := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, "/", nil)
			Expect(err).ToNot(HaveOccurred())

			res.List(recorder, req)
			response := recorder.Result()
			defer response.Body.Close()
			Expect(response.StatusCode).To(Equal(http.StatusOK))
			respData, err := io.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(respData)).To(Equal(expContents))
		})

	})
})
