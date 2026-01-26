package recover

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRecover(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Recover Suite")
}
