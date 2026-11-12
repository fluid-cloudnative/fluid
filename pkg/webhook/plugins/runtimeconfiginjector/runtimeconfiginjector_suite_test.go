package runtimeconfiginjector

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRuntimeconfiginjector(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Runtimeconfiginjector Suite")
}
