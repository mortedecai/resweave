package resweave_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

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
		DescribeTable("All methods should 405 initially",
			func(method string, path string, ctx context.Context) {
				req, err := http.NewRequest(method, path, nil)
				Expect(err).ToNot(HaveOccurred())
				_, _ = resultsMatch(http.StatusMethodNotAllowed, ctx, res, req)
			},
			Entry("GET", http.MethodGet, "/", contextWithURISegments([]string{})),
			Entry("POST", http.MethodPost, "/", contextWithURISegments([]string{})),
			Entry("GET", http.MethodGet, "/1/", contextWithURISegments([]string{"1"})),
			Entry("PUT", http.MethodPut, "/1/", contextWithURISegments([]string{"1"})),
			Entry("PATCH", http.MethodPatch, "/1/", contextWithURISegments([]string{"1"})),
			Entry("DELETE", http.MethodDelete, "/1/", contextWithURISegments([]string{"1"})),
			Entry("NOSUCHMETHOD", "NOSUCHMETHOD", "/", contextWithURISegments([]string{""})),
		)
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

			res.HandleCall(contextWithURISegments([]string{}), recorder, req)
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

			res.SetList(func(_ context.Context, w http.ResponseWriter, _ *http.Request) {
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
			req2, err := http.NewRequest(http.MethodPost, "/", nil)
			Expect(err).ToNot(HaveOccurred())

			res.HandleCall(contextWithURISegments([]string{}), recorder, req)
			response := recorder.Result()
			defer response.Body.Close()
			Expect(response.StatusCode).To(Equal(http.StatusOK))
			respData, err := io.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(respData)).To(Equal(expContents))
			respData, err = resultsMatch(http.StatusMethodNotAllowed, contextWithURISegments([]string{}), res, req2)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(respData)).To(Equal(""))
		})
		It("should allow a Create function to be called", func() {
			var s *zap.SugaredLogger
			if l, err := zap.NewDevelopment(); err == nil {
				s = l.Sugar()
			} else {
				Expect(err).ToNot(HaveOccurred())
			}

			res.SetCreate(func(_ context.Context, w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusCreated)
				respBytes := []byte("{}")
				if bw, err := w.Write(respBytes); err != nil {
					s.Infow("Create", "WriteError", err, "BytesWritten", bw)
				} else {
					s.Debugw("Create", "BytesWritten", bw)
				}
			})
			expContents := "{}"
			req, err := http.NewRequest(http.MethodPost, "/", nil)
			Expect(err).ToNot(HaveOccurred())
			req2, err := http.NewRequest(http.MethodGet, "/", nil)
			Expect(err).ToNot(HaveOccurred())
			respData, err := resultsMatch(http.StatusCreated, contextWithURISegments([]string{}), res, req)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(respData)).To(Equal(expContents))
			respData, err = resultsMatch(http.StatusMethodNotAllowed, contextWithURISegments([]string{}), res, req2)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(respData)).To(Equal(""))
		})
		DescribeTable("should be possible to add all viable functions to an API resource",
			func(method string, path string, ctx context.Context, expStatus int) {
				res.SetCreate(func(_ context.Context, w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusCreated)
				})
				res.SetList(func(_ context.Context, w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusOK)
				})
				res.SetID(resweave.NumericID)
				res.SetFetch(func(_ context.Context, w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusTeapot)
				})
				res.SetDelete(func(_ context.Context, w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusNoContent)
				})
				res.SetUpdate(func(_ context.Context, w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusAlreadyReported)
				})
				req, err := http.NewRequest(method, path, nil)
				Expect(err).ToNot(HaveOccurred())
				_, _ = resultsMatch(expStatus, ctx, res, req)
			},
			Entry("LIST", http.MethodGet, "/", contextWithURISegments([]string{}), http.StatusOK),
			Entry("FETCH", http.MethodGet, "/21/", contextWithURISegments(strings.Split("/21/", "/")), http.StatusTeapot),
			Entry("CREATE", http.MethodPost, "/", contextWithURISegments([]string{}), http.StatusCreated),
			Entry("UPDATE - Put", http.MethodPut, "/21/", contextWithURISegments([]string{"21"}), http.StatusAlreadyReported),
			Entry("UPDATE - Patch", http.MethodPatch, "/21/", contextWithURISegments([]string{"21"}), http.StatusAlreadyReported),
			Entry("DELETE", http.MethodDelete, "/1/", contextWithURISegments([]string{("1")}), http.StatusNoContent),
			Entry("Unknown method", "NOSUCHMETHOD", "/", contextWithURISegments([]string{}), http.StatusMethodNotAllowed),
		)
		DescribeTable("should be possible to add all viable functions to, then remove them from, an API resource",
			func(method string, path string, ctx context.Context, expStatus int) {
				res.SetCreate(func(_ context.Context, w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusCreated)
				})
				res.SetList(func(_ context.Context, w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusOK)
				})
				res.SetID(resweave.NumericID)
				res.SetFetch(func(_ context.Context, w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusTeapot)
				})
				res.SetDelete(func(_ context.Context, w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusNoContent)
				})
				res.SetUpdate(func(_ context.Context, w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusAlreadyReported)
				})
				res.SetCreate(nil)
				res.SetList(nil)
				res.SetFetch(nil)
				res.SetUpdate(nil)
				res.SetDelete(nil)
				req, err := http.NewRequest(method, path, nil)
				Expect(err).ToNot(HaveOccurred())
				_, _ = resultsMatch(expStatus, ctx, res, req)

			},
			Entry("LIST", http.MethodGet, "/", contextWithURISegments([]string{}), http.StatusMethodNotAllowed),
			Entry("FETCH", http.MethodGet, "/21/", contextWithURISegments(strings.Split("/21/", "/")), http.StatusMethodNotAllowed),
			Entry("CREATE", http.MethodPost, "/", contextWithURISegments([]string{}), http.StatusMethodNotAllowed),
			Entry("UPDATE - Put", http.MethodPut, "/21/", contextWithURISegments([]string{"21"}), http.StatusMethodNotAllowed),
			Entry("UPDATE - Patch", http.MethodPatch, "/21/", contextWithURISegments([]string{"21"}), http.StatusMethodNotAllowed),
			Entry("DELETE", http.MethodDelete, "/1/", contextWithURISegments([]string{("1")}), http.StatusMethodNotAllowed),
			Entry("Unknown method", "NOSUCHMETHOD", "/", contextWithURISegments([]string{}), http.StatusMethodNotAllowed),
		)
		It("should allow a custom handler to be registered", func() {
			// Arrange
			handled := false
			res.SetHandler(func(_ resweave.ActionType, _ context.Context, _ http.ResponseWriter, _ *http.Request) {
				handled = true
			})

			// Act
			req, err := http.NewRequest(http.MethodPost, "/test", nil)
			Expect(err).ToNot(HaveOccurred())
			res.HandleCall(contextWithURISegments([]string{"test"}), nil, req)

			// Assert
			Expect(handled).To(BeTrue())
		})
	})

	var _ = Describe("Fetch & List Handling", func() {
		var ()
		DescribeTable("should return the correct status code",
			func(method string, path string, ctx context.Context, expStatus int) {
				res := resweave.NewAPI("users")
				res.SetID(resweave.NumericID)

				res.SetFetch(func(_ context.Context, w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusNoContent)
				})

				res.SetList(func(_ context.Context, w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusAccepted)
				})
				l, err := zap.NewDevelopment()
				Expect(err).ToNot(HaveOccurred())
				logger := l.Sugar()
				res.SetLogger(logger, true)
				req, err := http.NewRequest(method, path, nil)
				Expect(err).ToNot(HaveOccurred())
				_, _ = resultsMatch(expStatus, ctx, res, req)
			},
			Entry("LIST", http.MethodGet, "users/", contextWithURISegments(strings.Split("users/", "/")), http.StatusAccepted),
			Entry("FETCH /<id>", http.MethodGet, "users/1", contextWithURISegments(strings.Split("users/1", "/")), http.StatusNoContent),
			Entry("FETCH /<id>/", http.MethodGet, "users/1/", contextWithURISegments(strings.Split("users/1/", "/")), http.StatusNoContent),
			Entry("FETCH /<bad id>/", http.MethodGet, "users/a/", contextWithURISegments(strings.Split("users/a/", "/")), http.StatusNotFound),
		)
	})

	var _ = Describe("GetIDValue", func() {
		entries := []struct {
			description string
			outcome     string
			ctx         context.Context
			expIDStr    string
			errMatcher  func(error)
		}{
			{
				description: "empty context",
				outcome:     "should error due to no ID",
				ctx:         context.Background(),
				expIDStr:    "",
				errMatcher:  func(err error) { Expect(err).To(MatchError(resweave.ErrIDNotFound)) },
			},
			{
				description: "no ID",
				outcome:     "should error due to no ID",
				ctx:         context.WithValue(context.Background(), resweave.Key("foo"), "bar"),
				expIDStr:    "",
				errMatcher:  func(err error) { Expect(err).To(MatchError(resweave.ErrIDNotFound)) },
			},
			{
				description: "ID exists",
				outcome:     "should return the ID value no error",
				ctx:         context.WithValue(context.Background(), resweave.Key("id_foo"), "1234"),
				expIDStr:    "1234",
				errMatcher:  func(err error) { Expect(err).ToNot(HaveOccurred()) },
			},
			{
				description: "Invalid ID",
				outcome:     "should return the ID value even if it is invalid",
				ctx:         context.WithValue(context.Background(), resweave.Key("id_foo"), "Not A Number"),
				expIDStr:    "Not A Number",
				errMatcher:  func(err error) { Expect(err).ToNot(HaveOccurred()) },
			},
		}

		for _, e := range entries {
			entry := e
			Context(entry.description, func() {
				It(entry.outcome, func() {
					res := resweave.NewAPI("foo")
					idStr, err := res.GetIDValue(entry.ctx)
					Expect(idStr).To(Equal(entry.expIDStr))
					entry.errMatcher(err)
				})
			})
		}
	})

	var _ = Describe("AddSubResource", func() {
		It("should be possible to add a sub-resource", func() {
			res := resweave.NewAPI("foo")
			subRes := resweave.NewAPI("bar")
			err := res.AddSubResource(subRes)
			Expect(err).ToNot(HaveOccurred())
		})
		It("should error if the sub-resource is nil", func() {
			res := resweave.NewAPI("foo")
			err := res.AddSubResource(nil)
			Expect(err).To(MatchError(resweave.ErrNilResource))
		})
		It("should error if the sub-resource is already added", func() {
			res := resweave.NewAPI("foo")
			subRes := resweave.NewAPI("bar")
			err := res.AddSubResource(subRes)
			Expect(err).ToNot(HaveOccurred())
			err = res.AddSubResource(subRes)
			Expect(err).To(MatchError(resweave.ErrResourceAlreadyExists))
		})
	})

	var _ = Describe("AddInstancedSubResource", func() {
		It("should be possible to add a sub-resource", func() {
			res := resweave.NewAPI("foo")
			subRes := resweave.NewAPI("bar")
			err := res.AddInstancedSubResource(subRes)
			Expect(err).ToNot(HaveOccurred())
		})
		It("should error if the sub-resource is nil", func() {
			res := resweave.NewAPI("foo")
			err := res.AddInstancedSubResource(nil)
			Expect(err).To(MatchError(resweave.ErrNilResource))
		})
		It("should error if the sub-resource is already added", func() {
			res := resweave.NewAPI("foo")
			subRes := resweave.NewAPI("bar")
			err := res.AddInstancedSubResource(subRes)
			Expect(err).ToNot(HaveOccurred())
			err = res.AddInstancedSubResource(subRes)
			Expect(err).To(MatchError(resweave.ErrInstancedResourceAlreadyExists))
		})
	})
})

func resultsMatch(expStatusCode int, inputContext context.Context, res resweave.Resource, req *http.Request) ([]byte, error) {
	recorder := httptest.NewRecorder()
	res.HandleCall(inputContext, recorder, req)
	response := recorder.Result()
	defer response.Body.Close()
	Expect(response.StatusCode).To(Equal(expStatusCode))
	return io.ReadAll(response.Body)
}

func contextWithURISegments(segments []string) context.Context {
	return context.WithValue(context.Background(), resweave.KeyURISegments, resweave.ResourceNames(segments))
}
