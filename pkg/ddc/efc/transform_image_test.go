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

package efc

import (
	corev1 "k8s.io/api/core/v1"
	"testing"
)

func TestParseMasterImage(t *testing.T) {
	engine := &EFCEngine{}
	image, tag, imagePullPolicy, ref := engine.parseMasterImage("", "", "", []corev1.LocalObjectReference{})
	if image != "registry.cn-zhangjiakou.aliyuncs.com/nascache/efc-master" ||
		tag != "latest" || imagePullPolicy != "IfNotPresent" || len(ref) != 0 {
		t.Errorf("unexpected err")
	}
}

func TestParseFuseImage(t *testing.T) {
	engine := &EFCEngine{}
	image, tag, imagePullPolicy, ref := engine.parseFuseImage("", "", "", []corev1.LocalObjectReference{})
	if image != "registry.cn-zhangjiakou.aliyuncs.com/nascache/efc-fuse" ||
		tag != "latest" || imagePullPolicy != "IfNotPresent" || len(ref) != 0 {
		t.Errorf("unexpected err")
	}
}

func TestParseWorkerImage(t *testing.T) {
	engine := &EFCEngine{}
	image, tag, imagePullPolicy, ref := engine.parseWorkerImage("", "", "", []corev1.LocalObjectReference{})
	if image != "registry.cn-zhangjiakou.aliyuncs.com/nascache/efc-worker" ||
		tag != "latest" || imagePullPolicy != "IfNotPresent" || len(ref) != 0 {
		t.Errorf("unexpected err")
	}
}

func TestParseInitFuseImage(t *testing.T) {
	engine := &EFCEngine{}
	image, tag, imagePullPolicy, ref := engine.parseInitFuseImage("", "", "", []corev1.LocalObjectReference{})
	if image != "registry.cn-zhangjiakou.aliyuncs.com/nascache/init-alifuse" ||
		tag != "latest" || imagePullPolicy != "IfNotPresent" || len(ref) != 0 {
		t.Errorf("unexpected err")
	}
}
