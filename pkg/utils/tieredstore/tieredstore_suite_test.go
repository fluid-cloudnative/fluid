package tieredstore

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTieredstore(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tieredstore Suite")
}
