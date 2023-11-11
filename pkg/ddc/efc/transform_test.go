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
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/net"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestEFCEngine_transform(t *testing.T) {
	var tests = []struct {
		runtime *datav1alpha1.EFCRuntime
		dataset *datav1alpha1.Dataset
	}{
		{&datav1alpha1.EFCRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.EFCRuntimeSpec{
				Fuse: datav1alpha1.EFCFuseSpec{},
				Worker: datav1alpha1.EFCCompTemplateSpec{
					Replicas: 2,
				},
			},
		}, &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{
					{
						MountPoint: "nfs://abcd-abc67.cn-zhangjiakou.nas.aliyuncs.com:/test-fluid-3/",
					},
				},
			},
		},
		},
	}
	for _, test := range tests {
		testObjs := []runtime.Object{}
		testObjs = append(testObjs, test.runtime.DeepCopy())
		testObjs = append(testObjs, test.dataset.DeepCopy())

		client := fake.NewFakeClientWithScheme(testScheme, testObjs...)
		engine := EFCEngine{
			name:      "test",
			namespace: "fluid",
			Client:    client,
			Log:       fake.NullLogger(),
			runtime:   test.runtime,
		}
		ctrl.SetLogger(zap.New(func(o *zap.Options) {
			o.Development = true
		}))
		err := portallocator.SetupRuntimePortAllocator(client, &net.PortRange{Base: 10, Size: 100}, "bitmap", GetReservedPorts)
		if err != nil {
			t.Fatal(err.Error())
		}
		_, err = engine.transform(test.runtime)
		if err != nil {
			t.Errorf("error %v", err)
		}
	}
}
