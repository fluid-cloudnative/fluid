package alluxio

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestAlluxio runs the test suite for the Alluxio package.
func TestAlluxio(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Alluxio Suite", Label("alluxio"))
}
