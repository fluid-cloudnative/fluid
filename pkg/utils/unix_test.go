package utils

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SplitSchemaAddr", func() {
	DescribeTable("should parse address correctly",
		func(addr string, wantProtocol string, wantAddr string) {
			gotProtocol, gotAddr := SplitSchemaAddr(addr)
			Expect(gotProtocol).To(Equal(wantProtocol))
			Expect(gotAddr).To(Equal(wantAddr))
		},
		Entry("unix protocol", "unix:///foo/bar", "unix", "/foo/bar"),
		Entry("tcp protocol", "tcp://127.0.0.1:8088", "tcp", "127.0.0.1:8088"),
		Entry("default protocol", "127.0.0.1:3456", "tcp", "127.0.0.1:3456"),
	)
})
