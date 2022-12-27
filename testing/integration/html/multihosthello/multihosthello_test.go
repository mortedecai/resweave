package main_test

import (
	"io"
	"net/http"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Multi Host Hello", func() {
	var _ = Describe("Default", func() {
		It("should be possible to receive the default hello world index page", func() {
			data, err := os.ReadFile("html/default/index.html")
			Expect(err).ToNot(HaveOccurred())
			expContents := string(data)
			resp, err := http.Get("http://multihosthello/")
			Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()
			respData, err := io.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(respData)).To(Equal(expContents))
		})
	})
	var _ = Describe("Named", func() {
		It("should be possible to receive the named hello world index page", func() {
			data, err := os.ReadFile("html/caHost/index.html")
			Expect(err).ToNot(HaveOccurred())
			expContents := string(data)
			resp, err := http.Get("http://mortedecai-ca")
			Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()
			respData, err := io.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(respData)).To(Equal(expContents))
		})
		It("a non-listed host should use the default host", func() {
			data, err := os.ReadFile("html/default/index.html")
			Expect(err).ToNot(HaveOccurred())
			expContents := string(data)
			resp, err := http.Get("http://mortedecai/")
			Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()
			respData, err := io.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(respData)).To(Equal(expContents))
		})

	})
})
