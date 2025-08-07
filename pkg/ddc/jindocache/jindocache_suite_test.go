package jindocache_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestJindocache(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Jindocache Suite")
}
