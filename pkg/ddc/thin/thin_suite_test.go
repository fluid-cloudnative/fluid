package thin_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestThin(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Thin Suite")
}
