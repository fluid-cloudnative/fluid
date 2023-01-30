package base

import "time"

// MetadataSyncResult describes result for asynchronous metadata sync
type MetadataSyncResult struct {
	Done      bool
	StartTime time.Time
	UfsTotal  string
	FileNum   string
	Err       error
}

// SafeClose closes the metadataSyncResultChannel but ignores panic when the channel is already closed.
// Returns true if the channel is already closed.
func SafeClose(ch chan MetadataSyncResult) (closed bool) {
	if ch == nil {
		return
	}
	defer func() {
		if recover() != nil {
			closed = true
		}
	}()

	close(ch)
	return false
}

// SafeSend sends result to the metadataSyncResultChannel but ignores panic when the channel is already closed
// Returns true if the channel is already closed.
func SafeSend(ch chan MetadataSyncResult, result MetadataSyncResult) (closed bool) {
	if ch == nil {
		return
	}
	defer func() {
		if recover() != nil {
			closed = true
		}
	}()

	ch <- result
	return false
}
