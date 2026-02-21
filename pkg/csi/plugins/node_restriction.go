package plugins

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
	return &NodeRestriction{}, nil
}
