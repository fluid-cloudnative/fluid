package poststart

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPoststart(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Poststart Suite")
}
