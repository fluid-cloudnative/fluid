package updatedbconf

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestUpdatedbconf(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Updatedbconf Suite")
}
