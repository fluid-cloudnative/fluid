/*
  Copyright 2026 The Fluid Authors.

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

package engine

import (
	"context"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type contextAwareCacheClient struct {
	client.Client
	updateCount int
}

func (c *contextAwareCacheClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return c.Client.Get(ctx, key, obj, opts...)
}

func (c *contextAwareCacheClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return c.Client.Create(ctx, obj, opts...)
}

func (c *contextAwareCacheClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	c.updateCount++
	return c.Client.Update(ctx, obj, opts...)
}

func TestCreateConfigMapInRuntimeClassWithCanceledContext(t *testing.T) {
	scheme := newCacheEngineTestScheme(t)
	baseClient := fake.NewFakeClientWithScheme(scheme)
	testClient := &contextAwareCacheClient{Client: baseClient}
	engine := &CacheEngine{Client: testClient, name: "demo", namespace: "default"}
	resources := &datav1alpha1.RuntimeExtraResources{
		ConfigMaps: []datav1alpha1.ConfigMapRuntimeExtraResource{{Name: "extra", Data: map[string]string{"key": "value"}}},
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := engine.createConfigMapInRuntimeClass(ctx, resources, nil); err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestCreateConfigMapInRuntimeClassCreatesMissingConfigMap(t *testing.T) {
	scheme := newCacheEngineTestScheme(t)
	baseClient := fake.NewFakeClientWithScheme(scheme)
	engine := &CacheEngine{Client: baseClient, name: "demo", namespace: "default"}
	resources := &datav1alpha1.RuntimeExtraResources{
		ConfigMaps: []datav1alpha1.ConfigMapRuntimeExtraResource{{Name: "extra", Data: map[string]string{"key": "value"}}},
	}

	if err := engine.createConfigMapInRuntimeClass(context.Background(), resources, nil); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	created := &corev1.ConfigMap{}
	if err := baseClient.Get(context.Background(), types.NamespacedName{Name: "extra", Namespace: "default"}, created); err != nil {
		t.Fatalf("expected configmap to be created, got %v", err)
	}
	if got := created.Data["key"]; got != "value" {
		t.Fatalf("expected copied data value, got %q", got)
	}
}

func TestCreateRuntimeValueConfigMapCreatesMissingConfigMap(t *testing.T) {
	scheme := newCacheEngineTestScheme(t)
	runtimeObj := newCacheRuntimeForConfigMapTest()
	runtimeClass := newCacheRuntimeClassForConfigMapTest()
	dataset := newDatasetForConfigMapTest()
	baseClient := fake.NewFakeClientWithScheme(scheme, runtimeObj, runtimeClass, dataset)
	engine := &CacheEngine{Client: baseClient, name: "demo", namespace: "default"}

	if err := engine.createRuntimeValueConfigMap(context.Background(), runtimeObj, nil); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	created := &corev1.ConfigMap{}
	if err := baseClient.Get(context.Background(), types.NamespacedName{Name: engine.getRuntimeConfigConfigMapName(), Namespace: "default"}, created); err != nil {
		t.Fatalf("expected runtime value configmap to be created, got %v", err)
	}
	if _, ok := created.Data[engine.getRuntimeConfigFileName()]; !ok {
		t.Fatalf("expected runtime value config data key %q", engine.getRuntimeConfigFileName())
	}
}

func TestSyncRuntimeValueConfigMapWithCanceledContext(t *testing.T) {
	scheme := newCacheEngineTestScheme(t)
	runtimeObj := newCacheRuntimeForConfigMapTest()
	runtimeClass := newCacheRuntimeClassForConfigMapTest()
	dataset := newDatasetForConfigMapTest()
	testClient := &contextAwareCacheClient{Client: fake.NewFakeClientWithScheme(scheme, runtimeObj, runtimeClass, dataset)}
	engine := &CacheEngine{Client: testClient, name: "demo", namespace: "default"}
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	ctx := cruntime.ReconcileRequestContext{Context: canceledCtx}

	if err := engine.syncRuntimeValueConfigMap(ctx, runtimeObj); err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestSyncRuntimeValueConfigMapSkipsUnchangedData(t *testing.T) {
	scheme := newCacheEngineTestScheme(t)
	runtimeObj := newCacheRuntimeForConfigMapTest()
	runtimeClass := newCacheRuntimeClassForConfigMapTest()
	dataset := newDatasetForConfigMapTest()

	baseEngine := &CacheEngine{Client: fake.NewFakeClientWithScheme(scheme, runtimeObj, runtimeClass, dataset), name: "demo", namespace: "default"}
	data, err := baseEngine.generateRuntimeConfigData(runtimeObj)
	if err != nil {
		t.Fatalf("failed to generate runtime config data: %v", err)
	}
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: baseEngine.getRuntimeConfigConfigMapName(), Namespace: "default"},
		Data:       data,
	}

	testClient := &contextAwareCacheClient{Client: fake.NewFakeClientWithScheme(scheme, runtimeObj, runtimeClass, dataset, configMap)}
	engine := &CacheEngine{Client: testClient, name: "demo", namespace: "default"}
	ctx := cruntime.ReconcileRequestContext{Context: context.Background()}

	if err := engine.syncRuntimeValueConfigMap(ctx, runtimeObj); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if testClient.updateCount != 0 {
		t.Fatalf("expected unchanged data to skip update, got %d updates", testClient.updateCount)
	}
}

func newCacheEngineTestScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := datav1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	return scheme
}

func newCacheRuntimeForConfigMapTest() *datav1alpha1.CacheRuntime {
	return &datav1alpha1.CacheRuntime{
		TypeMeta: metav1.TypeMeta{APIVersion: "data.fluid.io/v1alpha1", Kind: datav1alpha1.CacheRuntimeKind},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo",
			Namespace: "default",
			UID:       types.UID("demo-uid"),
		},
		Spec: datav1alpha1.CacheRuntimeSpec{
			RuntimeClassName: "test-class",
			Master:           datav1alpha1.CacheRuntimeMasterSpec{RuntimeComponentCommonSpec: datav1alpha1.RuntimeComponentCommonSpec{Disabled: true}},
			Worker:           datav1alpha1.CacheRuntimeWorkerSpec{RuntimeComponentCommonSpec: datav1alpha1.RuntimeComponentCommonSpec{Disabled: true}},
			Client:           datav1alpha1.CacheRuntimeClientSpec{RuntimeComponentCommonSpec: datav1alpha1.RuntimeComponentCommonSpec{Disabled: true}},
		},
	}
}

func newCacheRuntimeClassForConfigMapTest() *datav1alpha1.CacheRuntimeClass {
	return &datav1alpha1.CacheRuntimeClass{
		ObjectMeta: metav1.ObjectMeta{Name: "test-class"},
	}
}

func newDatasetForConfigMapTest() *datav1alpha1.Dataset {
	return &datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{Name: "demo", Namespace: "default"},
		Spec: datav1alpha1.DatasetSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadOnlyMany},
			Mounts: []datav1alpha1.Mount{
				{
					Name:       "hbase",
					MountPoint: "local:///data",
					Path:       "/data",
					ReadOnly:   true,
					Shared:     true,
				},
			},
		},
		Status: datav1alpha1.DatasetStatus{Runtimes: []datav1alpha1.Runtime{{Name: "demo", Type: common.CacheRuntime}}},
	}
}
