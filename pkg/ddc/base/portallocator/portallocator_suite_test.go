package portallocator_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPortallocator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Portallocator Suite")
}
