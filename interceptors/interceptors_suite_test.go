package interceptors_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestInterceptors(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Interceptors Suite")
}
