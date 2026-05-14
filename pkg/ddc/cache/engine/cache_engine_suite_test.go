package engine

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCacheEngine(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cache Engine Suite", Label("cache_engine"))
}
