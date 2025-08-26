package juicefs_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestJuicefs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Juicefs Suite")
}
