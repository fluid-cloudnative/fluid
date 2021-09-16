package common

import (
	"github.com/pkg/errors"
	"log"
	"os"
	"strconv"
)

const (
	EnvCriticalFusePodEnabled = "CRITICAL_FUSE_POD"
)

var criticalFusePodEnabled bool

func init() {
	if strVal, exist := os.LookupEnv(EnvCriticalFusePodEnabled); exist {
		if boolVal, err := strconv.ParseBool(strVal); err != nil {
			panic(errors.Wrapf(err, "can't parse env %s to bool", EnvCriticalFusePodEnabled))
		} else {
			criticalFusePodEnabled = boolVal
		}
	}
	log.Printf("Using %s = %v\n", EnvCriticalFusePodEnabled, criticalFusePodEnabled)
}

func CriticalFusePodEnabled() bool {
	return criticalFusePodEnabled
}
