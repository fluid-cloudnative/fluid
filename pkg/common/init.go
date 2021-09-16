package common

import (
	"github.com/pkg/errors"
	"log"
	"os"
	"strconv"
)

const (
	DefaultInitImage    = "registry.cn-hangzhou.aliyuncs.com/fluid/init-users:v0.3.0-1467caa"
	EnvPortCheckEnabled = "INIT_PORT_CHECK_ENABLED"
)

// The InitContainer to init the users for other Containers
type InitUsers struct {
	ImageInfo      `yaml:",inline"`
	EnvUsers       string `yaml:"envUsers"`
	Dir            string `yaml:"dir"`
	Enabled        bool   `yaml:"enabled,omitempty"`
	EnvTieredPaths string `yaml:"envTieredPaths"`
}

// InitPortCheck defines a init container reports port status usage
type InitPortCheck struct {
	ImageInfo    `yaml:",inline"`
	Enabled      bool   `yaml:"enabled,omitempty"`
	PortsToCheck string `yaml:"portsToCheck,omitempty"`
}

var initPortCheckEnabled = false

func init() {
	if strVal, exist := os.LookupEnv(EnvPortCheckEnabled); exist {
		if boolVal, err := strconv.ParseBool(strVal); err != nil {
			panic(errors.Wrapf(err, "can't parse %s to bool", EnvPortCheckEnabled))
		} else {
			initPortCheckEnabled = boolVal
		}
	}
	log.Printf("Using %s = %v\n", EnvPortCheckEnabled, initPortCheckEnabled)
}

func PortCheckEnabled() bool {
	return initPortCheckEnabled
}
