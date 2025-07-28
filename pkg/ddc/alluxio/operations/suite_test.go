package operations_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAlluxioFileUtilsSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AlluxioFileUtils Suite")
}
