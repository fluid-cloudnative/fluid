package mountpropagationinjector

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMountPropagationInjector(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MountPropagationInjector Suite")
}
