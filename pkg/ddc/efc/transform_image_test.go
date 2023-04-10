package efc

import (
	"testing"
)

func TestParseMasterImage(t *testing.T) {
	engine := &EFCEngine{}
	image, tag, imagePullPolicy := engine.parseMasterImage("", "", "")
	if image != "registry.cn-zhangjiakou.aliyuncs.com/nascache/efc-master" ||
		tag != "update" || imagePullPolicy != "IfNotPresent" {
		t.Errorf("unexpected err")
	}
}

func TestParseFuseImage(t *testing.T) {
	engine := &EFCEngine{}
	image, tag, imagePullPolicy := engine.parseFuseImage("", "", "")
	if image != "registry.cn-zhangjiakou.aliyuncs.com/nascache/efc-fuse" ||
		tag != "update" || imagePullPolicy != "IfNotPresent" {
		t.Errorf("unexpected err")
	}
}

func TestParseWorkerImage(t *testing.T) {
	engine := &EFCEngine{}
	image, tag, imagePullPolicy := engine.parseWorkerImage("", "", "")
	if image != "registry.cn-zhangjiakou.aliyuncs.com/nascache/efc-worker" ||
		tag != "update" || imagePullPolicy != "IfNotPresent" {
		t.Errorf("unexpected err")
	}
}

func TestParseInitFuseImage(t *testing.T) {
	engine := &EFCEngine{}
	image, tag, imagePullPolicy := engine.parseInitFuseImage("", "", "")
	if image != "registry.cn-zhangjiakou.aliyuncs.com/nascache/init-alifuse" ||
		tag != "update" || imagePullPolicy != "IfNotPresent" {
		t.Errorf("unexpected err")
	}
}
