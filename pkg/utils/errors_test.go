package utils

import (
	"fmt"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

func TestLoggingErrorExceptConflict(t *testing.T) {
	logger := fake.NullLogger()
	result := LoggingErrorExceptConflict(logger,
		apierrors.NewConflict(schema.GroupResource{},
			"test",
			fmt.Errorf("the object has been modified; please apply your changes to the latest version and try again")),
		"Failed to setup worker",
		types.NamespacedName{
			Namespace: "test",
			Name:      "test",
		})
	if result != nil {
		t.Errorf("Expected error result is null, but got %v", result)
	}
	result = LoggingErrorExceptConflict(logger,
		apierrors.NewNotFound(schema.GroupResource{}, "test"),
		"Failed to setup worker", types.NamespacedName{
			Namespace: "test",
			Name:      "test",
		})
	if result == nil {
		t.Errorf("Expected error result is not null, but got %v", result)
	}
}
