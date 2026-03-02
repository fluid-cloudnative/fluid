package prefernodeswithoutcache

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPrefernodeswithoutcache(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Prefernodeswithoutcache Suite")
}
