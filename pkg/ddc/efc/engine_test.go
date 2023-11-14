/*
  Copyright 2022 The Fluid Authors.

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
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

var (
	testScheme *runtime.Scheme
)

func init() {
	testScheme = runtime.NewScheme()
	_ = v1.AddToScheme(testScheme)
	_ = datav1alpha1.AddToScheme(testScheme)
	_ = appsv1.AddToScheme(testScheme)
}

func TestBuild(t *testing.T) {
	var namespace = v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "fluid",
		},
	}
	testObjs := []runtime.Object{}
	testObjs = append(testObjs, namespace.DeepCopy())

	var dataset = datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase",
			Namespace: "fluid",
		},
	}
	testObjs = append(testObjs, dataset.DeepCopy())

	var runtime = datav1alpha1.EFCRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase",
			Namespace: "fluid",
		},
		Spec: datav1alpha1.EFCRuntimeSpec{
			Fuse: datav1alpha1.EFCFuseSpec{
				CleanPolicy: datav1alpha1.OnDemandCleanPolicy,
			},
		},
		Status: datav1alpha1.RuntimeStatus{
			CacheStates: map[common.CacheStateName]string{
				common.Cached: "true",
			},
		},
	}
	testObjs = append(testObjs, runtime.DeepCopy())

	var sts = appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase-worker",
			Namespace: "fluid",
		},
	}
	testObjs = append(testObjs, sts.DeepCopy())
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	var ctx = cruntime.ReconcileRequestContext{
		NamespacedName: types.NamespacedName{
			Name:      "hbase",
			Namespace: "fluid",
		},
		Client:      client,
		Log:         fake.NullLogger(),
		RuntimeType: common.EFCRuntime,
		Runtime:     &runtime,
	}

	engine, err := Build("testId", ctx)
	if err != nil || engine == nil {
		t.Errorf("fail to exec the build function with the eror %v", err)
	}

	var errCtx = cruntime.ReconcileRequestContext{
		NamespacedName: types.NamespacedName{
			Name:      "hbase",
			Namespace: "fluid",
		},
		Client:      client,
		Log:         fake.NullLogger(),
		RuntimeType: common.EFCRuntime,
		Runtime:     nil,
	}

	got, err := Build("testId", errCtx)
	if err == nil {
		t.Errorf("expect err, but no err got %v", got)
	}

	var errCtx2 = cruntime.ReconcileRequestContext{
		NamespacedName: types.NamespacedName{
			Name:      "hbase2",
			Namespace: "fluid",
		},
		Client:      client,
		Log:         fake.NullLogger(),
		RuntimeType: common.EFCRuntime,
		Runtime:     &runtime,
	}

	got2, err2 := Build("testId", errCtx2)
	if err2 == nil {
		t.Errorf("expect err, but no err got %v", got2)
	}

	var errCtx3 = cruntime.ReconcileRequestContext{
		NamespacedName: types.NamespacedName{
			Name:      "hbase",
			Namespace: "fluid",
		},
		Client:      client,
		Log:         fake.NullLogger(),
		RuntimeType: common.EFCRuntime,
		Runtime:     &datav1alpha1.JindoRuntime{},
	}

	got3, err3 := Build("testId", errCtx3)
	if err3 == nil {
		t.Errorf("expect err, but no err got %v", got3)
	}
}
