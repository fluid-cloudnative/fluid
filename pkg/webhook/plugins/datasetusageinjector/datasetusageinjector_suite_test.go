package datasetusageinjector

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDatasetusageinjector(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Datasetusageinjector Suite")
}
