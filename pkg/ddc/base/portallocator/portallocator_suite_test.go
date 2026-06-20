package portallocator_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestPortallocator runs the port allocator test suite
// Description:
//
//	This function serves as the entry point for the Ginkgo test framework to execute all test cases
//	related to the port allocator. It registers the Ginkgo fail handler via RegisterFailHandler
//	and starts the test suite using RunSpecs.
//
// Parameters:
//
//	t *testing.T - The standard Go testing object used to control test execution flow
//
// Returns: no return value
func TestPortallocator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Portallocator Suite")
}
