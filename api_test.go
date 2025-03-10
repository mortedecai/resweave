package resweave_test

import (
	"context"
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
		It("should return a 405 for all methods initially", func() {
			testData := []struct {
				method string
				path   string
			}{
				{http.MethodGet, "/"},
				{http.MethodPost, "/"},
				{http.MethodPut, "/"},
				{http.MethodPatch, "/"},
				{http.MethodDelete, "/"},
				{"NOSUCHMETHOD", "/"},
			}
			for _, v := range testData {
				req, err := http.NewRequest(v.method, v.path, nil)
				Expect(err).ToNot(HaveOccurred())
				resultsMatch(http.StatusMethodNotAllowed, res, req)
			}

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

			res.HandleCall(context.TODO(), recorder, req)
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

			res.HandleCall(context.TODO(), recorder, req)
			response := recorder.Result()
			defer response.Body.Close()
			Expect(response.StatusCode).To(Equal(http.StatusOK))
			respData, err := io.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(respData)).To(Equal(expContents))
			respData, err = resultsMatch(http.StatusMethodNotAllowed, res, req2)
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
			respData, err := resultsMatch(http.StatusCreated, res, req)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(respData)).To(Equal(expContents))
			respData, err = resultsMatch(http.StatusMethodNotAllowed, res, req2)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(respData)).To(Equal(""))
		})
		It("should be possible to add all viable functions to an API resource", func() {
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
			testData := []struct {
				method    string
				path      string
				expStatus int
			}{
				{http.MethodGet, "/", http.StatusOK},
				{http.MethodGet, "/21/", http.StatusTeapot},
				{http.MethodPost, "/", http.StatusCreated},
				/* TODO:  Uncomment as these methods are added */
				// {http.MethodPut, "/"},
				// {http.MethodPatch, "/"},
				{http.MethodDelete, "/", http.StatusNoContent},
				{"NOSUCHMETHOD", "/", http.StatusMethodNotAllowed},
			}
			for _, v := range testData {
				req, err := http.NewRequest(v.method, v.path, nil)
				Expect(err).ToNot(HaveOccurred())
				resultsMatch(v.expStatus, res, req)
			}

		})
		It("should be possible to add all and then delete all viable functions to an API resource", func() {
			res.SetCreate(func(_ context.Context, w http.ResponseWriter, req *http.Request) {
				if req.Method != http.MethodPost {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
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
			res.SetUpdate(func(_ context.Context, w http.ResponseWriter, req *http.Request) {
				if req.Method != http.MethodPatch && req.Method != http.MethodPut {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
				w.WriteHeader(http.StatusAccepted)
			})

			testData := []struct {
				method    string
				path      string
				expStatus int
			}{
				{http.MethodGet, "/", http.StatusOK},
				{http.MethodGet, "/21/", http.StatusTeapot},
				{http.MethodPost, "/", http.StatusCreated},
				{http.MethodPut, "/21/", http.StatusAccepted},
				{http.MethodPatch, "/21/", http.StatusAccepted},
				{http.MethodDelete, "/", http.StatusNoContent},
				{"NOSUCHMETHOD", "/", http.StatusMethodNotAllowed},
			}
			for _, v := range testData {
				req, err := http.NewRequest(v.method, v.path, nil)
				Expect(err).ToNot(HaveOccurred())
				resultsMatch(v.expStatus, res, req)
			}
			res.SetCreate(nil)
			res.SetList(nil)
			res.SetFetch(nil)
			res.SetUpdate(nil)
			res.SetDelete(nil)
			testData = []struct {
				method    string
				path      string
				expStatus int
			}{
				{http.MethodGet, "/", http.StatusMethodNotAllowed},
				{http.MethodGet, "/21/", http.StatusMethodNotAllowed},
				{http.MethodPost, "/", http.StatusMethodNotAllowed},
				{http.MethodPut, "/21/", http.StatusMethodNotAllowed},
				{http.MethodPatch, "/21/", http.StatusMethodNotAllowed},
				{http.MethodDelete, "/", http.StatusMethodNotAllowed},
				{"NOSUCHMETHOD", "/", http.StatusMethodNotAllowed},
			}
			for _, v := range testData {
				req, err := http.NewRequest(v.method, v.path, nil)
				Expect(err).ToNot(HaveOccurred())
				resultsMatch(v.expStatus, res, req)
			}
		})
		It("should allow a custom handler to be registered", func() {
			// Arrange
			handled := false
			res.SetHandler(func(_ resweave.ActionType, _ context.Context, _ http.ResponseWriter, _ *http.Request) {
				handled = true
			})

			// Act
			req, err := http.NewRequest(http.MethodPost, "/test", nil)
			Expect(err).ToNot(HaveOccurred())
			res.HandleCall(context.TODO(), nil, req)

			// Assert
			Expect(handled).To(BeTrue())
		})
	})

	var _ = Describe("Fetch & List Handling", func() {
		var (
			res    resweave.APIResource
			logger *zap.SugaredLogger
		)
		BeforeEach(func() {

			res = resweave.NewAPI("users")
			res.SetID(resweave.NumericID)

			res.SetFetch(func(_ context.Context, w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			})

			res.SetList(func(_ context.Context, w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusAccepted)
			})
			l, err := zap.NewDevelopment()
			Expect(err).ToNot(HaveOccurred())
			logger = l.Sugar()
			res.SetLogger(logger, true)
		})
		It("should return the correct status code", func() {
			testData := []struct {
				method    string
				path      string
				expStatus int
			}{
				{http.MethodGet, res.Name().String() + "/", http.StatusAccepted},
				{http.MethodGet, res.Name().String() + "/1", http.StatusNoContent},
				{http.MethodGet, res.Name().String() + "/1/", http.StatusNoContent},
				{http.MethodGet, res.Name().String() + "/a/", http.StatusNotFound},
			}
			for _, v := range testData {
				req, err := http.NewRequest(v.method, v.path, nil)
				Expect(err).ToNot(HaveOccurred())
				resultsMatch(v.expStatus, res, req)
			}
		})
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
})

func resultsMatch(expStatusCode int, res resweave.Resource, req *http.Request) ([]byte, error) {
	recorder := httptest.NewRecorder()
	res.HandleCall(context.TODO(), recorder, req)
	response := recorder.Result()
	defer response.Body.Close()
	Expect(response.StatusCode).To(Equal(expStatusCode))
	return io.ReadAll(response.Body)
}
