package volume

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestVolume(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Volume Suite")
}
