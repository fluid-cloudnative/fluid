package alluxio

import (
	"fmt"
	"time"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
)

func (e *AlluxioEngine) Upgrade() (modified bool, err error) {
	opID := e.name + "-upgrade"
	// Use a fake context for internal op
	ctx := e.Helper.Context

	state, _ := e.GetCurrentState(opID)
	if state != base.StateIdle && state != base.StateCompleted {
		return false, fmt.Errorf("busy")
	}

	_ = e.TransitionState(ctx, opID, base.StateExecuting, "Upgrading")
	time.Sleep(10 * time.Millisecond)
	_ = e.TransitionState(ctx, opID, base.StateCompleted, "Success")
	
	return true, nil
}