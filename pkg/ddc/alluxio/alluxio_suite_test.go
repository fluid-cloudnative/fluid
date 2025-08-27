package alluxio_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAlluxio(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Alluxio Suite", Label("alluxio"))
}
