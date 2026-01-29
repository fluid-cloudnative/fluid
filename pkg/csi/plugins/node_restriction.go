package plugins

import "k8s.io/klog/v2"

// NodeRestriction represents node-level CSI constraints
type NodeRestriction struct {
	CSIDisabled bool
	MaxVolumes  int
}

// NodeRestrictionChecker fetches node restrictions
type NodeRestrictionChecker interface {
	GetRestriction(nodeName string) (*NodeRestriction, error)
}

// noopRestrictionChecker allows everything (scaffolding only)
type noopRestrictionChecker struct{}

// NewNoopRestrictionChecker returns a checker that never restricts
func NewNoopRestrictionChecker() NodeRestrictionChecker {
	return &noopRestrictionChecker{}
}

func (c *noopRestrictionChecker) GetRestriction(nodeName string) (*NodeRestriction, error) {
	klog.V(4).Infof("noop node restriction checker for node %s", nodeName)
	return &NodeRestriction{}, nil
}
