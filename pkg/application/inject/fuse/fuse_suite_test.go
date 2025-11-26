package fuse_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFuse(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Application Fuse Injection Suite")
}
