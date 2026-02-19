package requirenodewithfuse

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRequirenodewithfuse(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Requirenodewithfuse Suite")
}
