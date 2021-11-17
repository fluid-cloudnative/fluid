package utils

import (
	"fmt"
	"testing"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	testLog "sigs.k8s.io/controller-runtime/pkg/log"
)

func TestLoggingErrorExceptConflict(t *testing.T) {
	logger := testLog.NullLogger{}
	result := LoggingErrorExceptConflict(logger, apierrors.NewConflict(schema.GroupResource{}, "test", fmt.Errorf("the object has been modified; please apply your changes to the latest version and try again")))
	if result != nil {
		t.Errorf("Expected error result is null, but got %v", result)
	}
	result = LoggingErrorExceptConflict(logger, apierrors.NewNotFound(schema.GroupResource{}, "test"))
	if result == nil {
		t.Errorf("Expected error result is not null, but got %v", result)
	}
}
