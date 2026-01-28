package mutating_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMutating(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mutating Suite")
}
