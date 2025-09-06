package kubeclient_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestKubeclient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pacakge Util - Kubeclient Suite")
}
