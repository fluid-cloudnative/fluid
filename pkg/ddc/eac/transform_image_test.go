/*
Copyright 2023 The Fluid Author.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
