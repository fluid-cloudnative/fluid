package utils

import "errors"

type NeedRetryWithApiReaderError struct {
	ErrMsg string
}

var _ error = &NeedRetryWithApiReaderError{}

// Error implements the Error interface.
func (e *NeedRetryWithApiReaderError) Error() string {
	return e.ErrMsg
}

// IsNeedRetryWithApiReaderError also recognizes the error when it has been wrapped with %w.
func IsNeedRetryWithApiReaderError(err error) bool {
	var target *NeedRetryWithApiReaderError
	return errors.As(err, &target)
}

func NewNeedRetryWithApiReaderError(err error) *NeedRetryWithApiReaderError {
	if err == nil {
		return nil
	}
	return &NeedRetryWithApiReaderError{
		ErrMsg: err.Error(),
	}
}

// RejectPodError is returned for a failure the pod spec itself causes, so retrying cannot
// help: the pod must be rejected rather than admitted without its mutation.
type RejectPodError struct {
	ErrMsg string
}

var _ error = &RejectPodError{}

// Error implements the Error interface.
func (e *RejectPodError) Error() string {
	return e.ErrMsg
}

// IsRejectPodError also recognizes the error when it has been wrapped with %w.
func IsRejectPodError(err error) bool {
	var target *RejectPodError
	return errors.As(err, &target)
}

func NewRejectPodError(err error) *RejectPodError {
	if err == nil {
		return nil
	}
	return &RejectPodError{
		ErrMsg: err.Error(),
	}
}
