package mountpropagationinjector

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMountpropagationinjector(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mountpropagationinjector Suite")
}
