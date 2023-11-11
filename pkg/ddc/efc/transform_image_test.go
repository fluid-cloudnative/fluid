/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
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
