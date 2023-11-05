package assertions

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/smarty/assertions/internal/go-render/render"
)

type Serializer interface {
	serialize(expected, actual any, message string) string
	serializeDetailed(expected, actual any, message string) string
}

type failureSerializer struct{}

func (self *failureSerializer) serializeDetailed(expected, actual any, message string) string {
	if index := strings.Index(message, " Diff:"); index > 0 {
		message = message[:index]
	}
	view := FailureView{
		Message:  message,
		Expected: render.Render(expected),
		Actual:   render.Render(actual),
	}
	serialized, _ := json.Marshal(view)
	return string(serialized)
}

func (self *failureSerializer) serialize(expected, actual any, message string) string {
	if index := strings.Index(message, " Diff:"); index > 0 {
		message = message[:index]
	}
	view := FailureView{
		Message:  message,
		Expected: fmt.Sprintf("%+v", expected),
		Actual:   fmt.Sprintf("%+v", actual),
	}
	serialized, _ := json.Marshal(view)
	return string(serialized)
}

func newSerializer() *failureSerializer {
	return &failureSerializer{}
}

///////////////////////////////////////////////////////////////////////////////

// FailureView is also declared in github.com/smarty/goconvey/convey/reporting.
// The json struct tags should be equal in both declarations.
type FailureView struct {
	Message  string `json:"Message"`
	Expected string `json:"Expected"`
	Actual   string `json:"Actual"`
}

///////////////////////////////////////////////////////

// noopSerializer just gives back the original message. This is useful when we are using
// the assertions from a context other than the GoConvey Web UI, that requires the JSON
// structure provided by the failureSerializer.
type noopSerializer struct{}

func (self *noopSerializer) serialize(expected, actual any, message string) string {
	return message
}
func (self *noopSerializer) serializeDetailed(expected, actual any, message string) string {
	return message
}
