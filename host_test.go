package resweave

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
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
		It("should have a nil logger eventually", func() {
			Expect(caHost.Logger()).To(BeNil())
		})
	})
	Describe("Serving Basics", func() {
		It("should receive all ids in instanced sub-resources", func() {
			path := "/resource/id-123/other"
			segments := ResourceNames(strings.Split(path, "/"))
			// split sets the first elment to an empty string if the path starts with a slash
			res := NewAPI(segments[1])
			res.SetID(`id-[0-9]+`)
			subRes := NewAPI(segments[3])
			res.AddChildResource(subRes)
			Expect(caHost.AddResource(res)).ToNot(HaveOccurred())
			handlerCalled := false
			subRes.SetHandler(func(_ ActionType, ctx context.Context, w http.ResponseWriter, _ *http.Request) {
				handlerCalled = true
				actSegs, valid := ctx.Value(KeyURISegments).([]ResourceName)
				Expect(valid).To(BeTrue())
				Expect(actSegs).To(Equal([]ResourceName{}))
				Expect(ctx.Value(Key(fmt.Sprintf("id_%s", res.Name())))).To(Equal("id-123"))
				Expect(ctx.Value(Key(fmt.Sprintf("id_%s", subRes.Name())))).To(BeNil())
			})
			req := httptest.NewRequest(http.MethodGet, path, nil)
			caHost.Serve(httptest.NewRecorder(), req)
			Expect(handlerCalled).To(BeTrue())
		})
		It("should store the path segments from idx 0 if the resource is not found", func() {
			path := "/resource/id-123/other"
			segments := ResourceNames(strings.Split(path, "/"))
			res := NewAPI("")
			resources := NewAPI(segments[1])
			resources.SetID(`id-[0-9]+`)
			subRes := NewAPI(segments[3])
			resources.AddChildResource(subRes)
			res.AddResource(resources)
			Expect(caHost.AddResource(res)).ToNot(HaveOccurred())
			handlerCalled := false
			subRes.SetHandler(func(_ ActionType, ctx context.Context, w http.ResponseWriter, _ *http.Request) {
				handlerCalled = true
				actSegs, valid := ctx.Value(KeyURISegments).([]ResourceName)
				Expect(valid).To(BeTrue())
				Expect(actSegs).To(Equal([]ResourceName{}))
				Expect(ctx.Value(Key(fmt.Sprintf("id_%s", res.Name())))).To(BeNil())
				Expect(ctx.Value(Key(fmt.Sprintf("id_%s", resources.Name())))).To(Equal("id-123"))
				Expect(ctx.Value(Key(fmt.Sprintf("id_%s", subRes.Name())))).To(BeNil())
			})
			req := httptest.NewRequest(http.MethodGet, path, nil)
			caHost.Serve(httptest.NewRecorder(), req)
			Expect(handlerCalled).To(BeTrue())

		})
		It("should be able to handle several levels of direction", func() {
			path := "/users/123/posts/456/comments/789/replies"
			segments := ResourceNames(strings.Split(path, "/"))
			usersRes := NewAPI(segments[1])
			usersRes.SetID(NumericID)
			postsRes := NewAPI(segments[3])
			postsRes.SetID(NumericID)
			commentsRes := NewAPI(segments[5])
			commentsRes.SetID(NumericID)
			repliesRes := NewAPI(segments[7])
			usersRes.AddChildResource(postsRes)
			postsRes.AddChildResource(commentsRes)
			commentsRes.AddChildResource(repliesRes)
			caHost.AddResource(usersRes)

			handlerCalled := false
			repliesRes.SetHandler(func(_ ActionType, ctx context.Context, w http.ResponseWriter, _ *http.Request) {
				handlerCalled = true
				actSegs, valid := ctx.Value(KeyURISegments).([]ResourceName)
				Expect(valid).To(BeTrue())
				Expect(actSegs).To(Equal([]ResourceName{}))
				Expect(ctx.Value(Key(fmt.Sprintf("id_%s", usersRes.Name())))).To(Equal("123"))
				Expect(ctx.Value(Key(fmt.Sprintf("id_%s", postsRes.Name())))).To(Equal("456"))
				Expect(ctx.Value(Key(fmt.Sprintf("id_%s", commentsRes.Name())))).To(Equal("789"))
				Expect(ctx.Value(Key(fmt.Sprintf("id_%s", repliesRes.Name())))).To(BeNil())
			})
			req := httptest.NewRequest(http.MethodGet, path, nil)
			caHost.Serve(httptest.NewRecorder(), req)
			Expect(handlerCalled).To(BeTrue())
		})
		It("should store an empty path segments slice if no segments exist", func() {
			path := "/"
			res := NewAPI("")
			Expect(caHost.AddResource(res)).ToNot(HaveOccurred())
			handlerCalled := false
			res.SetHandler(func(_ ActionType, ctx context.Context, w http.ResponseWriter, _ *http.Request) {
				handlerCalled = true
				actSegs, valid := ctx.Value(KeyURISegments).([]ResourceName)
				Expect(valid).To(BeTrue())
				Expect(actSegs).To(Equal([]ResourceName{}))
			})
			req := httptest.NewRequest(http.MethodGet, path, nil)
			caHost.Serve(httptest.NewRecorder(), req)
			Expect(handlerCalled).To(BeTrue())
		})
	})
	Describe("API Usage", func() {
		const (
			usersName = ResourceName("users")
			usersPath = "/users"
		)
		var (
			usersRes Resource
		)

		BeforeEach(func() {
			usersRes = NewAPI(usersName)
		})
		It("should be possible to add a named API resource", func() {
			Expect(caHost.AddResource(usersRes)).ToNot(HaveOccurred())
		})
		It("should be possible to retrieve the resource after adding an unnamed resource", func() {
			Expect(caHost.AddResource(usersRes)).ToNot(HaveOccurred())
			Expect(caHost.TopLevelResourceCount()).To(Equal(1))
			res, found := caHost.GetResource(usersName)
			Expect(found).To(BeTrue())
			Expect(res).To(Equal(usersRes))
		})
		It("should return an error if two unnamed resources are added", func() {
			Expect(caHost.AddResource(usersRes)).ToNot(HaveOccurred())
			Expect(caHost.TopLevelResourceCount()).To(Equal(1))
			Expect(caHost.AddResource(usersRes)).To(HaveOccurred())
			Expect(caHost.AddResource(usersRes)).To(Equal(fmt.Errorf(FmtResourceAlreadyExists, usersRes.Name(), caHost.Name())))
			Expect(caHost.TopLevelResourceCount()).To(Equal(1))
		})
		It("should serve a named api resource correctly", func() {
			l, _ := zap.NewDevelopment()
			s := l.Sugar()
			caHost.SetLogger(s, true)
			Expect(caHost.AddResource(usersRes)).ToNot(HaveOccurred())
			expContents := "Hello, World!"
			usersRes.(*BaseAPIRes).SetList(func(_ context.Context, w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				respBytes := []byte(expContents)
				if bw, err := w.Write(respBytes); err != nil {
					s.Infow("List", "WriteError", err, "BytesWritten", bw)
				} else {
					s.Debugw("List", "BytesWritten", bw)
				}
			})

			recorder := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, usersPath, nil)
			Expect(err).ToNot(HaveOccurred())

			caHost.Serve(recorder, req)
			response := recorder.Result()
			defer response.Body.Close()
			Expect(response.StatusCode).To(Equal(http.StatusOK))
			respData, err := io.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(respData)).To(Equal(expContents))
		})

	})
	Describe("HTML Usage", func() {
		const (
			htmlDir  = "test/html/"
			htmlDir2 = "test/html2/"
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
			Expect(caHost.AddResource(htmlRes)).To(Equal(fmt.Errorf(FmtResourceAlreadyExists, htmlRes.Name(), caHost.Name())))
			Expect(caHost.TopLevelResourceCount()).To(Equal(1))
		})
		It("should serve an unnamed root resource correctly", func() {
			Expect(caHost.AddResource(NewHTML("", htmlDir))).ToNot(HaveOccurred())
			data, err := os.ReadFile(htmlDir + "index.html")
			Expect(err).ToNot(HaveOccurred())
			expContents := string(data)
			recorder := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, "/", nil)
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
			req, err := http.NewRequest(http.MethodGet, "/", nil)
			Expect(err).ToNot(HaveOccurred())

			caHost.Serve(recorder, req)
			response := recorder.Result()
			defer response.Body.Close()
			Expect(response.StatusCode).To(Equal(http.StatusNotFound))
		})
		It("should be possible to add two named resources to the root", func() {
			Expect(caHost.AddResource(NewHTML("", htmlDir))).ToNot(HaveOccurred())
			Expect(caHost.AddResource(NewHTML("two", htmlDir2))).ToNot(HaveOccurred())
			data, err := os.ReadFile(htmlDir + "index.html")
			Expect(err).ToNot(HaveOccurred())
			expContents := string(data)
			data2, err := os.ReadFile(htmlDir2 + "index.html")
			Expect(err).ToNot(HaveOccurred())
			expContents2 := string(data2)
			data3, err := os.ReadFile(htmlDir + "test.html")
			Expect(err).ToNot(HaveOccurred())
			expContents3 := string(data3)
			data4, err := os.ReadFile(htmlDir2 + "test.html")
			Expect(err).ToNot(HaveOccurred())
			expContents4 := string(data4)

			recorder := httptest.NewRecorder()
			recorder2 := httptest.NewRecorder()
			recorder3 := httptest.NewRecorder()
			recorder4 := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, "/", nil)
			Expect(err).ToNot(HaveOccurred())
			req2, err := http.NewRequest(http.MethodGet, "/two/", nil)
			Expect(err).ToNot(HaveOccurred())
			req3, err := http.NewRequest(http.MethodGet, "/test.html", nil)
			Expect(err).ToNot(HaveOccurred())
			req4, err := http.NewRequest(http.MethodGet, "/two/test.html", nil)
			Expect(err).ToNot(HaveOccurred())

			caHost.Serve(recorder, req)
			caHost.Serve(recorder2, req2)
			caHost.Serve(recorder3, req3)
			caHost.Serve(recorder4, req4)

			response := recorder.Result()
			defer response.Body.Close()
			Expect(response.StatusCode).To(Equal(http.StatusOK))
			response2 := recorder2.Result()
			defer response2.Body.Close()
			Expect(response2.StatusCode).To(Equal(http.StatusOK))
			response3 := recorder3.Result()
			defer response3.Body.Close()
			Expect(response3.StatusCode).To(Equal(http.StatusOK))
			response4 := recorder4.Result()
			defer response4.Body.Close()
			Expect(response4.StatusCode).To(Equal(http.StatusOK))

			respData, err := io.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(respData)).To(Equal(expContents))

			respData2, err := io.ReadAll(response2.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(respData2)).To(Equal(expContents2))

			respData3, err := io.ReadAll(response3.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(respData3)).To(Equal(expContents3))

			respData4, err := io.ReadAll(response4.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(respData4)).To(Equal(expContents4))
		})
		It("should be possible to set the logger non-recursively", func() {
			Expect(caHost.AddResource(NewHTML("", htmlDir))).ToNot(HaveOccurred())
			l, err := zap.NewProduction()
			Expect(err).ToNot(HaveOccurred())

			Expect(caHost.Logger()).To(BeNil())
			caHost.SetLogger(l.Sugar(), false)
			Expect(caHost.Logger()).ToNot(BeNil())
			Expect(caHost.(*host).resources[""].Logger()).To(BeNil())
			caHost.SetLogger(nil, false)
			Expect(caHost.Logger()).To(BeNil())
			Expect(caHost.(*host).resources[""].Logger()).To(BeNil())
		})

	})
})
