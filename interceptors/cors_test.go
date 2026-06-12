package interceptors_test

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mortedecai/resweave/interceptors"
)

var _ = Describe("Cors", func() {
	Describe("NewCORS", func() {
		var next http.Handler

		BeforeEach(func() {
			next = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
		})

		It("returns a non-nil handler without error", func() {
			handler, err := interceptors.NewCORS(next)
			Expect(err).ToNot(HaveOccurred())
			Expect(handler).ToNot(BeNil())
		})

		It("calls the next handler", func() {
			called := false
			handler, err := interceptors.NewCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
			}))
			Expect(err).ToNot(HaveOccurred())

			handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

			Expect(called).To(BeTrue())
		})

		It("applies provided options on each request", func() {
			handler, err := interceptors.NewCORS(next,
				interceptors.WithOrigin("https://example.com"),
				interceptors.WithMethods("GET", "POST"),
			)
			Expect(err).ToNot(HaveOccurred())

			for i := 0; i < 3; i++ {
				recorder := httptest.NewRecorder()
				handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
				Expect(recorder.Header().Get("Access-Control-Allow-Origin")).To(Equal("https://example.com"))
				Expect(recorder.Header().Get("Access-Control-Allow-Methods")).To(Equal("GET,POST"))
			}
		})
	})

	Describe("Options", func() {
		var (
			recorder *httptest.ResponseRecorder
			w        http.ResponseWriter
		)

		BeforeEach(func() {
			recorder = httptest.NewRecorder()
			w = recorder
		})

		Describe("WithOrigin", func() {
			It("sets Access-Control-Allow-Origin", func() {
				interceptors.WithOrigin("https://example.com")(&w)
				Expect(recorder.Header().Get("Access-Control-Allow-Origin")).To(Equal("https://example.com"))
			})

			It("returns a CORSOption", func() {
				var opt interceptors.CORSOption = interceptors.WithOrigin("https://example.com")
				Expect(opt(&w)).ToNot(BeNil())
				Expect(recorder.Header().Get("Access-Control-Allow-Origin")).To(Equal("https://example.com"))
			})
		})

		Describe("WithMethods", func() {
			It("sets Access-Control-Allow-Methods from a single method", func() {
				interceptors.WithMethods("GET")(&w)
				Expect(recorder.Header().Get("Access-Control-Allow-Methods")).To(Equal("GET"))
			})

			It("sets Access-Control-Allow-Methods from multiple methods", func() {
				interceptors.WithMethods("GET", "POST", "PUT", "DELETE")(&w)
				Expect(recorder.Header().Get("Access-Control-Allow-Methods")).To(Equal("GET,POST,PUT,DELETE"))
			})

			It("returns a CORSOption", func() {
				var opt interceptors.CORSOption = interceptors.WithMethods("GET", "POST")
				Expect(opt(&w)).ToNot(BeNil())
				Expect(recorder.Header().Get("Access-Control-Allow-Methods")).To(Equal("GET,POST"))
			})
		})

		Describe("WithHeaders", func() {
			It("sets Access-Control-Allow-Headers from a single header", func() {
				interceptors.WithHeaders("Content-Type")(&w)
				Expect(recorder.Header().Get("Access-Control-Allow-Headers")).To(Equal("Content-Type"))
			})

			It("sets Access-Control-Allow-Headers from multiple headers", func() {
				interceptors.WithHeaders("Content-Type", "Authorization")(&w)
				Expect(recorder.Header().Get("Access-Control-Allow-Headers")).To(Equal("Content-Type,Authorization"))
			})

			It("returns a CORSOption", func() {
				var opt interceptors.CORSOption = interceptors.WithHeaders("Content-Type", "Authorization")
				Expect(opt(&w)).ToNot(BeNil())
				Expect(recorder.Header().Get("Access-Control-Allow-Headers")).To(Equal("Content-Type,Authorization"))
			})
		})

		Describe("AllowCredentials", func() {
			It("sets Access-Control-Allow-Credentials to true", func() {
				interceptors.AllowCredentials("true")(&w)
				Expect(recorder.Header().Get("Access-Control-Allow-Credentials")).To(Equal("true"))
			})

			It("sets Access-Control-Allow-Credentials to false", func() {
				interceptors.AllowCredentials("false")(&w)
				Expect(recorder.Header().Get("Access-Control-Allow-Credentials")).To(Equal("false"))
			})

			It("returns a CORSOption", func() {
				var opt interceptors.CORSOption = interceptors.AllowCredentials("true")
				Expect(opt(&w)).ToNot(BeNil())
				Expect(recorder.Header().Get("Access-Control-Allow-Credentials")).To(Equal("true"))
			})
		})

		Describe("WithMaxAge", func() {
			It("sets Access-Control-Max-Age", func() {
				interceptors.WithMaxAge("3600")(&w)
				Expect(recorder.Header().Get("Access-Control-Max-Age")).To(Equal("3600"))
			})

			It("returns a CORSOption", func() {
				var opt interceptors.CORSOption = interceptors.WithMaxAge("3600")
				Expect(opt(&w)).ToNot(BeNil())
				Expect(recorder.Header().Get("Access-Control-Max-Age")).To(Equal("3600"))
			})
		})

		Describe("composing multiple options", func() {
			It("applies all options independently", func() {
				interceptors.WithOrigin("https://example.com")(&w)
				interceptors.WithMethods("GET", "POST")(&w)
				interceptors.WithHeaders("Content-Type", "Authorization")(&w)
				interceptors.AllowCredentials("true")(&w)
				interceptors.WithMaxAge("600")(&w)

				Expect(recorder.Header().Get("Access-Control-Allow-Origin")).To(Equal("https://example.com"))
				Expect(recorder.Header().Get("Access-Control-Allow-Methods")).To(Equal("GET,POST"))
				Expect(recorder.Header().Get("Access-Control-Allow-Headers")).To(Equal("Content-Type,Authorization"))
				Expect(recorder.Header().Get("Access-Control-Allow-Credentials")).To(Equal("true"))
				Expect(recorder.Header().Get("Access-Control-Max-Age")).To(Equal("600"))
			})
		})
	})
})
