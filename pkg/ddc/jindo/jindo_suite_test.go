package jindo_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestJindo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Jindo Suite")
}
