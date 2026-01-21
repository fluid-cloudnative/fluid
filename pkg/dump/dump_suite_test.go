package dump

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDump(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dump Suite")
}
