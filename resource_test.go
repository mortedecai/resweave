package resweave_test

import (
	"github.com/mortedecai/resweave"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Resources", func() {
	It("should be possible to create an unnamed HTML resource", func() {
		htmlRes := resweave.NewHTML("", htmlDir)
		Expect(htmlRes).ToNot(BeNil())
	})
})
