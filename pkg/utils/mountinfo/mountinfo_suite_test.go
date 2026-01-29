package mountinfo

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMountinfo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mountinfo Suite")
}
