package alluxio

import (
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

func TestAlluxioInPlaceUpgrade(t *testing.T) {
	// 1. Setup Engine
	engine := &AlluxioEngine{
		TemplateEngine: &base.TemplateEngine{
			Id:                              "test-runtime",
			Log:                             fake.NullLogger(),
			DefaultExtendedLifecycleManager: base.DefaultExtendedLifecycleManager{},
		},
		StateMachineManager: base.NewDefaultStateMachine(),
	}

	// 2. Run Upgrade (First time, should succeed)
	modified, err := engine.Upgrade()
	if err != nil {
		t.Errorf("Upgrade failed: %v", err)
	}
	if !modified {
		t.Error("Expected modified to be true")
	}

	// 3. Verify State is Completed
	opID := "test-runtime-upgrade"
	state, _ := engine.GetCurrentState(opID)
	if state != base.StateCompleted {
		t.Errorf("Expected state Completed, got %s", state)
	}

	// 4. Test logic when busy (Simulate Executing state manually)
	// Force set state to Executing
	_ = engine.TransitionState(engine.Context, opID, base.StateExecuting, "Force Busy")

	// Try Upgrade again - should fail because it's busy
	_, err = engine.Upgrade()
	if err == nil {
		t.Error("Expected error when upgrading busy runtime, got nil")
	}
}