package resweave

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

var _ = Describe("Logger", func() {
	var (
		lhnr LogHolder
		lhr  LogHolder
	)
	BeforeEach(func() {
		lhnr = NewLogholder("lhnr", nil)
		lhr = NewLogholder("lhr", recurse)
	})
	It("should be possible to create a new non-recursing logHolder", func() {
		Expect(lhnr).ToNot(BeNil())
		Expect(lhnr.Logger()).To(BeNil())
		Expect(lhnr.LoggerName()).To(Equal("lhnr"))
	})
	It("should not recurse when the non-recuring logHolder is called with true", func() {
		rcBefore := recurseCount
		lhnr.SetLogger(nil, true)
		rcAfter := recurseCount
		Expect(rcBefore).To(Equal(rcAfter))
	})
	It("should be possible to create a new recursing logHolder", func() {
		Expect(lhr).ToNot(BeNil())
		Expect(lhr.Logger()).To(BeNil())
		Expect(lhr.LoggerName()).To(Equal("lhr"))
	})
	It("should recurse when the non-recuring logHolder is called with true", func() {
		rcBefore := recurseCount
		lhr.SetLogger(nil, true)
		rcAfter := recurseCount
		Expect(rcBefore < rcAfter).To(BeTrue())
	})
})

var (
	recurseCount = 0
)

func recurse(l *zap.SugaredLogger) {
	recurseCount++
}
