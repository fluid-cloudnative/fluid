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

package alluxio

import (
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/net"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var dummy = func(client client.Client) (ports []int, err error) {
	return []int{20001, 20002, 20003}, nil
}

func TestShutdown(t *testing.T) {
	// runtimeInfoSpark tests destroy Worker in exclusive mode.
	runtimeInfoSpark, err := base.BuildRuntimeInfo("spark", "fluid", "alluxio", datav1alpha1.TieredStore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}
	runtimeInfoSpark.SetupWithDataset(&datav1alpha1.Dataset{
		Spec: datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ExclusiveMode},
	})

	// runtimeInfoSpark tests destroy Worker in shareMode mode.
	runtimeInfoHadoop, err := base.BuildRuntimeInfo("hadoop", "fluid", "alluxio", datav1alpha1.TieredStore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}
	runtimeInfoHadoop.SetupWithDataset(&datav1alpha1.Dataset{
		Spec: datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ShareMode},
	})
	nodeSelector := map[string]string{
		"node-select": "true",
	}
	runtimeInfoHadoop.SetupFuseDeployMode(true, nodeSelector)

	var nodeInputs = []*corev1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node-spark",
				Labels: map[string]string{
					"fluid.io/dataset-num":               "1",
					"fluid.io/s-alluxio-fluid-spark":     "true",
					"fluid.io/s-fluid-spark":             "true",
					"fluid.io/s-h-alluxio-d-fluid-spark": "5B",
					"fluid.io/s-h-alluxio-m-fluid-spark": "1B",
					"fluid.io/s-h-alluxio-t-fluid-spark": "6B",
					"fluid_exclusive":                    "fluid_spark",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node-share",
				Labels: map[string]string{
					"fluid.io/dataset-num":                "2",
					"fluid.io/s-alluxio-fluid-hadoop":     "true",
					"fluid.io/s-fluid-hadoop":             "true",
					"fluid.io/s-h-alluxio-d-fluid-hadoop": "5B",
					"fluid.io/s-h-alluxio-m-fluid-hadoop": "1B",
					"fluid.io/s-h-alluxio-t-fluid-hadoop": "6B",
					"fluid.io/s-alluxio-fluid-hbase":      "true",
					"fluid.io/s-fluid-hbase":              "true",
					"fluid.io/s-h-alluxio-d-fluid-hbase":  "5B",
					"fluid.io/s-h-alluxio-m-fluid-hbase":  "1B",
					"fluid.io/s-h-alluxio-t-fluid-hbase":  "6B",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node-hadoop",
				Labels: map[string]string{
					"fluid.io/dataset-num":                "1",
					"fluid.io/s-alluxio-fluid-hadoop":     "true",
					"fluid.io/s-fluid-hadoop":             "true",
					"fluid.io/s-h-alluxio-d-fluid-hadoop": "5B",
					"fluid.io/s-h-alluxio-m-fluid-hadoop": "1B",
					"fluid.io/s-h-alluxio-t-fluid-hadoop": "6B",
					"node-select":                         "true",
				},
			},
		},
	}

	testNodes := []runtime.Object{}
	for _, nodeInput := range nodeInputs {
		testNodes = append(testNodes, nodeInput.DeepCopy())
	}

	var runtimeInputs = []*datav1alpha1.AlluxioRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.AlluxioRuntimeSpec{
				Master: datav1alpha1.AlluxioCompTemplateSpec{
					NetworkMode: datav1alpha1.ContainerNetworkMode,
				},
				Worker: datav1alpha1.AlluxioCompTemplateSpec{
					NetworkMode: datav1alpha1.ContainerNetworkMode,
				},
			},
		}, {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hadoop",
				Namespace: "fluid",
			},
		},
	}

	for _, runtimeInput := range runtimeInputs {
		testNodes = append(testNodes, runtimeInput.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testNodes...)

	pr := net.ParsePortRangeOrDie("20000-21000")

	err = portallocator.SetupRuntimePortAllocator(nil, pr, "bitmap", dummy)
	if err != nil {
		t.Fatalf("failed to set up runtime port allocator due to %v", err)
	}

	var testCase = []struct {
		expectedWorkers  int32
		runtimeInfo      base.RuntimeInfoInterface
		wantedNodeNumber int32
		wantedNodeLabels map[string]map[string]string
	}{
		{
			expectedWorkers:  -1,
			runtimeInfo:      runtimeInfoSpark,
			wantedNodeNumber: 0,
			wantedNodeLabels: map[string]map[string]string{
				"test-node-spark": {},
				"test-node-share": {
					"fluid.io/dataset-num":                "2",
					"fluid.io/s-alluxio-fluid-hadoop":     "true",
					"fluid.io/s-fluid-hadoop":             "true",
					"fluid.io/s-h-alluxio-d-fluid-hadoop": "5B",
					"fluid.io/s-h-alluxio-m-fluid-hadoop": "1B",
					"fluid.io/s-h-alluxio-t-fluid-hadoop": "6B",
					"fluid.io/s-alluxio-fluid-hbase":      "true",
					"fluid.io/s-fluid-hbase":              "true",
					"fluid.io/s-h-alluxio-d-fluid-hbase":  "5B",
					"fluid.io/s-h-alluxio-m-fluid-hbase":  "1B",
					"fluid.io/s-h-alluxio-t-fluid-hbase":  "6B",
				},
				"test-node-hadoop": {
					"fluid.io/dataset-num":                "1",
					"fluid.io/s-alluxio-fluid-hadoop":     "true",
					"fluid.io/s-fluid-hadoop":             "true",
					"fluid.io/s-h-alluxio-d-fluid-hadoop": "5B",
					"fluid.io/s-h-alluxio-m-fluid-hadoop": "1B",
					"fluid.io/s-h-alluxio-t-fluid-hadoop": "6B",
					"node-select":                         "true",
				},
			},
		},
		{
			expectedWorkers:  -1,
			runtimeInfo:      runtimeInfoHadoop,
			wantedNodeNumber: 0,
			wantedNodeLabels: map[string]map[string]string{
				"test-node-spark": {},
				"test-node-share": {
					"fluid.io/dataset-num":               "1",
					"fluid.io/s-alluxio-fluid-hbase":     "true",
					"fluid.io/s-fluid-hbase":             "true",
					"fluid.io/s-h-alluxio-d-fluid-hbase": "5B",
					"fluid.io/s-h-alluxio-m-fluid-hbase": "1B",
					"fluid.io/s-h-alluxio-t-fluid-hbase": "6B",
				},
				"test-node-hadoop": {
					"node-select": "true",
				},
			},
		},
	}
	for _, test := range testCase {
		engine := &AlluxioEngine{Log: fake.NullLogger(), runtimeInfo: test.runtimeInfo}
		engine.Client = client
		engine.name = test.runtimeInfo.GetName()
		engine.namespace = test.runtimeInfo.GetNamespace()
		engine.Helper = ctrl.BuildHelper(engine.runtimeInfo, client, engine.Log)

		queryCacheStatusPatches := ApplyPrivateMethod(engine, "queryCacheStatus", func() (cacheStates, error) {
			return cacheStates{
				cacheCapacity:    "19.07MiB",
				cached:           "0.00B",
				cachedPercentage: "0.0%",
				cacheHitStates: cacheHitStates{
					bytesReadLocal:  20310917,
					bytesReadUfsAll: 32243712,
				},
			}, nil
		})

		invokeCleanCachePatches := ApplyPrivateMethod(engine, "invokeCleanCache", func(path string) error {
			return nil
		})

		destroyMasterPatches := ApplyPrivateMethod(engine, "destroyMaster", func() error {
			return nil
		})

		func() {
			defer destroyMasterPatches.Reset()
			defer invokeCleanCachePatches.Reset()
			defer queryCacheStatusPatches.Reset()
			err := engine.Shutdown()
			if err != nil {
				t.Fatalf("failed to engine.Shutdonw() due to: %v", err)
			}
		}()

		// if !(err == nil ||
		// 	strings.Contains(err.Error(), "executable file not found") ||
		// 	strings.Contains(err.Error(), "invalid configuration") ||
		// 	strings.Contains(err.Error(), "not found")) {
		// 	t.Errorf("fail to call the shutdown with the error %v", err)
		// }
	}
}
