package operations

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestOperations is the entry point for the Ginkgo test suite of the operations package.
// It registers the Gomega fail handler and runs all Ginkgo specs defined in the package
// under the suite name "Operations Suite".
// This function is invoked by the Go test runner via `go test`.
//
// Parameters:
//   - t (*testing.T): The standard Go testing object passed by the test runner.
func TestOperations(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Operations Suite")
}
