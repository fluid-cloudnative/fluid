package alluxio

import (
	"testing"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

func TestAlluxioLoadData(t *testing.T) {
	engine := &AlluxioEngine{
		DefaultExtendedLifecycleManager: base.DefaultExtendedLifecycleManager{},
		StateMachineManager:             base.NewDefaultStateMachine(),
		Log:                             fake.NullLogger(),
	}

	spec := &base.DataLoadSpec{Paths: []string{"/data"}}
	ctx := cruntime.ReconcileRequestContext{Name: "test"}
	
	res, err := engine.LoadData(ctx, spec)
	if err != nil {
		t.Errorf("LoadData failed: %v", err)
	}
	if res.Status.Phase != base.DataOperationPhaseCompleted {
		t.Error("Expected completed")
	}
}