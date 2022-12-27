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
	It("should be possible to create a set of resource names from a slice of strings", func() {
		s1 := []string{"one"}
		e1 := []resweave.ResourceName{resweave.ResourceName("one")}
		Expect(resweave.ResourceNames(s1)).To(Equal(e1))
		s2 := []string{"one", "two", "three", "four"}
		e2 := []resweave.ResourceName{
			resweave.ResourceName("one"),
			resweave.ResourceName("two"),
			resweave.ResourceName("three"),
			resweave.ResourceName("four"),
		}
		Expect(resweave.ResourceNames(s2)).To(Equal(e2))
	})
})
