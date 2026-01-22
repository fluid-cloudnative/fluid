package kubelet

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestKubelet(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kubelet Suite")
}
