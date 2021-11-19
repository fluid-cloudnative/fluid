package base

import (
	"context"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"

	fluiderrs "github.com/fluid-cloudnative/fluid/pkg/errors"
)

func TestLoggingErrorExceptConflict(t *testing.T) {

	engine := NewTemplateEngine(nil, "id", runtime.ReconcileRequestContext{
		Context: context.Background(),
		NamespacedName: types.NamespacedName{
			Namespace: "test",
			Name:      "test",
		},
		Log: log.NullLogger{},
	})

	err := engine.loggingErrorExceptConflict(fluiderrs.NewDeprecated(schema.GroupResource{Group: "", Resource: "test"}, types.NamespacedName{}), "test")
	if !fluiderrs.IsDeprecated(err) {
		t.Errorf("Failed to check deprecated error %v", err)
	}
}
