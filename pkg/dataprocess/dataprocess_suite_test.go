package dataprocess

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDataprocess(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dataprocess Suite")
}
