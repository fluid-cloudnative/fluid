package app

import (
	"github.com/fluid-cloudnative/fluid/pkg/webhook/scheduler/mutating"
)

func init() {
	addHandlers(mutating.HandlerMap)
	// addHandlers(validating.HandlerMap)
}
