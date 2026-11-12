package utils

import (
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Webhook error types", func() {
	cause := errors.New("configmap not found")

	It("should recognize a NeedRetryWithApiReaderError, also when wrapped", func() {
		err := NewNeedRetryWithApiReaderError(cause)
		Expect(IsNeedRetryWithApiReaderError(err)).To(BeTrue())
		Expect(IsNeedRetryWithApiReaderError(fmt.Errorf("plugin failed: %w", err))).To(BeTrue())
		Expect(IsRejectPodError(err)).To(BeFalse())
	})

	It("should recognize a RejectPodError, also when wrapped", func() {
		err := NewRejectPodError(cause)
		Expect(IsRejectPodError(err)).To(BeTrue())
		Expect(IsRejectPodError(fmt.Errorf("plugin failed: %w", err))).To(BeTrue())
		Expect(IsNeedRetryWithApiReaderError(err)).To(BeFalse())
	})

	It("should not match a plain error or nil", func() {
		Expect(IsNeedRetryWithApiReaderError(cause)).To(BeFalse())
		Expect(IsRejectPodError(cause)).To(BeFalse())
		Expect(IsNeedRetryWithApiReaderError(nil)).To(BeFalse())
		Expect(IsRejectPodError(nil)).To(BeFalse())
	})

	It("should return nil for a nil cause", func() {
		Expect(NewNeedRetryWithApiReaderError(nil)).To(BeNil())
		Expect(NewRejectPodError(nil)).To(BeNil())
	})
})
