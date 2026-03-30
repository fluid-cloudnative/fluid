package engine

import "fmt"

// TODO(cache runtime): Implement all functions using this function.
func newNotImplementError(name string) error {
	return fmt.Errorf("CacheRuntime does not implement '%s'", name)
}
