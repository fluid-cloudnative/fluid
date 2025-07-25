package utils

type NeedRetryWithApiReaderError struct {
	ErrMsg string
}

var _ error = &NeedRetryWithApiReaderError{}

// Error implements the Error interface.
func (e *NeedRetryWithApiReaderError) Error() string {
	return e.ErrMsg
}

func IsNeedRetryWithApiReaderError(err error) bool {
	if _, ok := err.(*NeedRetryWithApiReaderError); ok {
		return true
	}

	return false
}

func NewNeedRetryWithApiReaderError(err error) *NeedRetryWithApiReaderError {
	if err == nil {
		return nil
	}
	return &NeedRetryWithApiReaderError{
		ErrMsg: err.Error(),
	}
}
