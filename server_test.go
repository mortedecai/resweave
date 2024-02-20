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
	"go.uber.org/zap"
)

var _ = Describe("Server", func() {
	const (
		port = 8080
		// htmlDir re-declared b/c it is not in resweave_test package.
		htmlDir = "testing/html/"
	)

	Describe("Initialization", func() {
		var (
			s Server
		)
		BeforeEach(func() {
			s = NewServer(port).(*server)
		})
		It("should be possible to create a new server", func() {
			Expect(s).ToNot(BeNil())
			Expect(s.Port()).To(Equal(port))
			Expect(s.(*server).Logger()).To(BeNil())
			Expect(s.(*server).interceptor).ToNot(BeNil())
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
		It("should be possible to set a logger on the server only", func() {
			l, err := zap.NewProduction()
			Expect(err).ToNot(HaveOccurred())
			s.SetLogger(l.Sugar(), false)
			Expect(s.(*server).Logger()).ToNot(BeNil())
			Expect(s.(*server).hosts[""].Logger()).To(BeNil())
			s.SetLogger(nil, false)
			Expect(s.(*server).Logger()).To(BeNil())
			Expect(s.(*server).hosts[""].Logger()).To(BeNil())
		})
		It("should be possible to recursively set a logger on the server", func() {
			l, err := zap.NewProduction()
			Expect(err).ToNot(HaveOccurred())
			s.SetLogger(l.Sugar(), true)
			Expect(s.(*server).Logger()).ToNot(BeNil())
			Expect(s.(*server).hosts[""].Logger()).ToNot(BeNil())
			s.SetLogger(nil, true)
			Expect(s.(*server).Logger()).To(BeNil())
			Expect(s.(*server).hosts[""].Logger()).To(BeNil())
		})
		It("should be possible to add an interceptor", func() {
			ic1 := 0
			ic2 := 0

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			s.(*server).interceptor.ServeHTTP(recorder, req)
			Expect(ic1).To(BeZero())
			Expect(ic2).To(BeZero())

			s.AddInterceptor(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					ic1++
					next.ServeHTTP(w, r)
				})
			})
			recorder = httptest.NewRecorder()
			req = httptest.NewRequest(http.MethodGet, "/", nil)
			s.(*server).interceptor.ServeHTTP(recorder, req)
			Expect(ic1).To(Equal(1))
			Expect(ic2).To(BeZero())

			s.AddInterceptor(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					ic2++
					next.ServeHTTP(w, r)
				})
			})
			recorder = httptest.NewRecorder()
			req = httptest.NewRequest(http.MethodGet, "/", nil)
			l, err := zap.NewProduction()
			Expect(err).ToNot(HaveOccurred())
			s.SetLogger(l.Sugar(), true)
			// Necessary whitebox test:  s.(*server).setRequetIDInterceptor(s.interceptor) is called when executing Run()
			// To simulate the full request chain, it is necessary to call the following, rather than just interceptor.ServeHTTP
			srvr := s.(*server)
			srvr.setRequestIDInterceptor(srvr.interceptor).ServeHTTP(recorder, req)
			Expect(ic1).To(Equal(2))
			Expect(ic2).To(Equal(1))

		})
	})
	Describe("Host Names", func() {
		var (
			s Server
		)
		BeforeEach(func() {
			s = NewServer(port).(*server)
		})
		Describe("Default Host", func() {
			It("should be possible to add a resource to the default host", func() {
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
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				Expect(err).ToNot(HaveOccurred())

				s.(*server).Serve(recorder, req)
				response := recorder.Result()
				defer response.Body.Close()
				Expect(response.StatusCode).To(Equal(http.StatusOK))
				respData, err := io.ReadAll(response.Body)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(respData)).To(Equal(expContents))

			})
		})
		Describe("Named Host", func() {
			const (
				dotCAHost HostName = HostName("daniel-taylor.ca")
			)
			It("should be possible to add a named Host to the Server", func() {
				expHost, err := s.AddHost(dotCAHost)
				Expect(err).ToNot(HaveOccurred())
				Expect(expHost).ToNot(BeNil())
				host, found := s.GetHost(dotCAHost)
				Expect(found).To(BeTrue())
				Expect(host).To(Equal(expHost))
			})
			It("should not be possible to add the same named Host to the Server", func() {
				expHost, err := s.AddHost(dotCAHost)
				Expect(err).ToNot(HaveOccurred())
				Expect(expHost).ToNot(BeNil())
				oHost, err := s.AddHost(dotCAHost)
				Expect(err).To(HaveOccurred())
				Expect(oHost).To(BeNil())
			})
			It("should be possible to add a resource to a named host", func() {
				Expect(s.AddResource(NewHTML("", htmlDir))).ToNot(HaveOccurred())
				v, s := s.GetResource("")
				Expect(v).ToNot(BeNil())
				Expect(v.Name()).To(BeEquivalentTo(""))
				Expect(s).To(BeTrue())
			})
			It("should serve an unnamed root resource correctly", func() {
				h, err := s.AddHost(dotCAHost)
				Expect(err).ToNot(HaveOccurred())
				Expect(h.AddResource(NewHTML("", htmlDir))).ToNot(HaveOccurred())
				data, err := os.ReadFile(htmlDir + "index.html")
				Expect(err).ToNot(HaveOccurred())
				expContents := string(data)
				recorder := httptest.NewRecorder()
				recorder2 := httptest.NewRecorder()
				req, err := http.NewRequest(http.MethodGet, "https://daniel-taylor.ca:8080/", nil)
				Expect(err).ToNot(HaveOccurred())
				req2, err := http.NewRequest(http.MethodGet, "https://localhost:8080/", nil)
				Expect(err).ToNot(HaveOccurred())

				s.(*server).Serve(recorder, req)
				s.(*server).Serve(recorder2, req2)
				response := recorder.Result()
				response2 := recorder2.Result()
				defer response.Body.Close()
				defer response2.Body.Close()

				Expect(response.StatusCode).To(Equal(http.StatusOK))
				respData, err := io.ReadAll(response.Body)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(respData)).To(Equal(expContents))
				Expect(response2.StatusCode).To(Equal(http.StatusNotFound))
			})
		})
	})
})
