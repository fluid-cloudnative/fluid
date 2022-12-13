package eac

import (
	"testing"
)

func TestParseMasterImage(t *testing.T) {
	engine := &EACEngine{}
	image, tag, imagePullPolicy := engine.parseMasterImage("", "", "")
	if image != "registry.cn-zhangjiakou.aliyuncs.com/nasteam/eac-fluid-img" ||
		tag != "update" || imagePullPolicy != "IfNotPresent" {
		t.Errorf("unexpected err")
	}
}

func TestParseFuseImage(t *testing.T) {
	engine := &EACEngine{}
	image, tag, imagePullPolicy := engine.parseFuseImage("", "", "")
	if image != "registry.cn-zhangjiakou.aliyuncs.com/nasteam/eac-fluid-img" ||
		tag != "update" || imagePullPolicy != "IfNotPresent" {
		t.Errorf("unexpected err")
	}
}

func TestParseWorkerImage(t *testing.T) {
	engine := &EACEngine{}
	image, tag, imagePullPolicy := engine.parseWorkerImage("", "", "")
	if image != "registry.cn-zhangjiakou.aliyuncs.com/nasteam/eac-worker-img" ||
		tag != "update" || imagePullPolicy != "IfNotPresent" {
		t.Errorf("unexpected err")
	}
}

func TestParseInitFuseImage(t *testing.T) {
	engine := &EACEngine{}
	image, tag, imagePullPolicy := engine.parseInitFuseImage("", "", "")
	if image != "registry.cn-zhangjiakou.aliyuncs.com/nasteam/init-alifuse" ||
		tag != "update" || imagePullPolicy != "IfNotPresent" {
		t.Errorf("unexpected err")
	}
}
