package main_test

import (
	"io"
	"net/http"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Multi Root Hello", func() {
	var _ = Describe("Default", func() {
		It("should be possible to receive the default hello world index page", func() {
			data, err := os.ReadFile("html/folderOne/index.html")
			Expect(err).ToNot(HaveOccurred())
			expContents := string(data)
			resp, err := http.Get("http://multiroothello/")
			Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()
			respData, err := io.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(respData)).To(Equal(expContents))
		})
		It("should be possible to receive the goodbye page for the first folder", func() {
			data, err := os.ReadFile("html/folderOne/goodbye.html")
			Expect(err).ToNot(HaveOccurred())
			expContents := string(data)
			resp, err := http.Get("http://multiroothello/goodbye.html")
			Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()
			respData, err := io.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(respData)).To(Equal(expContents))
		})

		It("should be possible to get the index from the second directory", func() {
			data, err := os.ReadFile("html/folderTwo/index.html")
			Expect(err).ToNot(HaveOccurred())
			expContents := string(data)
			resp, err := http.Get("http://multiroothello/two/")
			Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()
			respData, err := io.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(respData)).To(Equal(expContents))
		})
		It("should be possible to receive the goodbye page for the second folder", func() {
			data, err := os.ReadFile("html/folderTwo/goodbye.html")
			Expect(err).ToNot(HaveOccurred())
			expContents := string(data)
			resp, err := http.Get("http://multiroothello/two/goodbye.html")
			Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()
			respData, err := io.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(respData)).To(Equal(expContents))
		})

	})
})
