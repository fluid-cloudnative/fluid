package nodeaffinitywithcache

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestNodeaffinitywithcache(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Nodeaffinitywithcache Suite")
}
