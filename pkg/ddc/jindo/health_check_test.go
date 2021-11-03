package jindo

import (
	"testing"
)

func TestCheckRuntimeHealthy(t *testing.T) {
	engine := &JindoEngine{}
	err := engine.CheckRuntimeHealthy()
	if err != nil {
		t.Errorf("check runtime healthy failed,err:%s", err.Error())
	}
}
