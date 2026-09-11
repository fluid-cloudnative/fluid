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
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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
	if err := baseClient.Get(context.Background(), types.NamespacedName{Name: common.GetCacheRuntimeConfigConfigMapName(engine.name), Namespace: "default"}, created); err != nil {
		t.Fatalf("expected runtime value configmap to be created, got %v", err)
	}
	if _, ok := created.Data[engine.getRuntimeConfigFileName()]; !ok {
		t.Fatalf("expected runtime value config data key %q", engine.getRuntimeConfigFileName())
	}
}

func TestGenerateRuntimeConfigDataWithCanceledContext(t *testing.T) {
	scheme := newCacheEngineTestScheme(t)
	runtimeObj := newCacheRuntimeForConfigMapTest()
	runtimeClass := newCacheRuntimeClassForConfigMapTest()
	dataset := newDatasetForConfigMapTest()
	testClient := &contextAwareCacheClient{Client: fake.NewFakeClientWithScheme(scheme, runtimeObj, runtimeClass, dataset)}
	engine := &CacheEngine{Client: testClient, name: "demo", namespace: "default"}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := engine.generateRuntimeConfigData(ctx, runtimeObj); err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
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
	data, err := baseEngine.generateRuntimeConfigData(context.Background(), runtimeObj)
	if err != nil {
		t.Fatalf("failed to generate runtime config data: %v", err)
	}
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: common.GetCacheRuntimeConfigConfigMapName(baseEngine.name), Namespace: "default"},
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
		Topology: &datav1alpha1.RuntimeTopology{
			Master: &datav1alpha1.RuntimeComponentDefinition{},
			Worker: &datav1alpha1.RuntimeComponentDefinition{},
			Client: &datav1alpha1.RuntimeComponentDefinition{},
		},
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
func TestGenerateRuntimeConfigDataWithMissingComponentTopology(t *testing.T) {
	testCases := map[string]struct {
		topology *datav1alpha1.RuntimeTopology
		enable   func(runtime *datav1alpha1.CacheRuntime)
	}{
		"master topology undefined": {
			topology: &datav1alpha1.RuntimeTopology{
				Worker: &datav1alpha1.RuntimeComponentDefinition{},
				Client: &datav1alpha1.RuntimeComponentDefinition{},
			},
			enable: func(r *datav1alpha1.CacheRuntime) { r.Spec.Master.Disabled = false },
		},
		"worker topology undefined": {
			topology: &datav1alpha1.RuntimeTopology{
				Master: &datav1alpha1.RuntimeComponentDefinition{},
				Client: &datav1alpha1.RuntimeComponentDefinition{},
			},
			enable: func(r *datav1alpha1.CacheRuntime) { r.Spec.Worker.Disabled = false },
		},
		"client topology undefined": {
			topology: &datav1alpha1.RuntimeTopology{
				Master: &datav1alpha1.RuntimeComponentDefinition{},
				Worker: &datav1alpha1.RuntimeComponentDefinition{},
			},
			enable: func(r *datav1alpha1.CacheRuntime) { r.Spec.Client.Disabled = false },
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			scheme := newCacheEngineTestScheme(t)
			runtimeObj := newCacheRuntimeForConfigMapTest()
			tc.enable(runtimeObj)
			runtimeClass := newCacheRuntimeClassForConfigMapTest()
			runtimeClass.Topology = tc.topology
			dataset := newDatasetForConfigMapTest()
			baseClient := fake.NewFakeClientWithScheme(scheme, runtimeObj, runtimeClass, dataset)
			engine := &CacheEngine{Client: baseClient, name: "demo", namespace: "default"}

			if _, err := engine.generateRuntimeConfigData(context.Background(), runtimeObj); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
func TestGenerateRuntimeConfigDataWithoutAnyComponent(t *testing.T) {
	testCases := map[string]struct {
		topology *datav1alpha1.RuntimeTopology
	}{
		"topology is nil":                {topology: nil},
		"topology declares no component": {topology: &datav1alpha1.RuntimeTopology{}},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			scheme := newCacheEngineTestScheme(t)
			runtimeObj := newCacheRuntimeForConfigMapTest()
			runtimeClass := newCacheRuntimeClassForConfigMapTest()
			runtimeClass.Topology = tc.topology
			dataset := newDatasetForConfigMapTest()
			baseClient := fake.NewFakeClientWithScheme(scheme, runtimeObj, runtimeClass, dataset)
			engine := &CacheEngine{Client: baseClient, name: "demo", namespace: "default"}

			_, err := engine.generateRuntimeConfigData(context.Background(), runtimeObj)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), "at least one component should be defined") {
				t.Fatalf("unexpected error message: %v", err)
			}
		})
	}
}
func TestGenerateDataLoadValueFileWithNilTopology(t *testing.T) {
	scheme := newCacheEngineTestScheme(t)
	runtimeObj := newCacheRuntimeForConfigMapTest()
	runtimeClass := &datav1alpha1.CacheRuntimeClass{
		ObjectMeta: metav1.ObjectMeta{Name: "test-class"},
		Topology:   nil,
		DataOperationSpecs: []datav1alpha1.DataOperationSpec{
			{Name: "DataLoad", Command: []string{"/usr/local/bin/dataload"}},
		},
	}
	dataset := newDatasetForConfigMapTest()
	dataload := &datav1alpha1.DataLoad{
		ObjectMeta: metav1.ObjectMeta{Name: "test-dataload", Namespace: "default"},
		Spec: datav1alpha1.DataLoadSpec{
			Dataset: datav1alpha1.TargetDataset{Name: "demo", Namespace: "default"},
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, runtimeObj, runtimeClass, dataset, dataload)
	engine := &CacheEngine{Client: fakeClient, name: "demo", namespace: "default"}
	ctx := cruntime.ReconcileRequestContext{Client: fakeClient}

	_, err := engine.generateDataLoadValueFile(ctx, dataload)
	if err == nil {
		t.Fatal("expected error when topology is nil, got nil")
	}
}

// --- runtime.sh generation ---------------------------------------------------------
//
// runtime.sh is the second face of the runtime config, added for cache systems that have
// no FUSE client and read their configuration through a native SDK instead (Mooncake).
// Those consumers `source` the file, so the tests below check the two properties a
// sourced file must have and that a JSON blob does not give for free: it must survive
// the shell (quoting), and it must be byte-stable across reconciles (no ConfigMap churn).

// newClientlessRuntimeForConfigMapTest is the shape #6176 targets: master and worker run,
// the client (the FUSE daemon) is switched off, and the app reads runtime.sh instead of
// mounting a FUSE volume.
func newClientlessRuntimeForConfigMapTest() *datav1alpha1.CacheRuntime {
	runtimeObj := newCacheRuntimeForConfigMapTest()
	runtimeObj.Spec.Master.Disabled = false
	runtimeObj.Spec.Master.Replicas = 1
	runtimeObj.Spec.Worker.Disabled = false
	runtimeObj.Spec.Worker.Replicas = 3
	runtimeObj.Spec.Client.Disabled = true
	return runtimeObj
}

func newEngineForConfigMapTest(t *testing.T, objs ...runtime.Object) *CacheEngine {
	t.Helper()
	return &CacheEngine{Client: fake.NewFakeClientWithScheme(newCacheEngineTestScheme(t), objs...), name: "demo", namespace: "default"}
}

// sourceRuntimeSh runs the generated script through a real shell and returns the value the
// shell ended up with for each requested variable. Values are printed separated by a
// marker rather than newlines, so a value containing spaces or newlines survives.
func TestRuntimeShExportsAbsentComponentsUnderNounset(t *testing.T) {
	// A topology may leave a component out. Its variables are still exported, from a zero
	// value, so that a consumer running under "set -u" can read any of them without first
	// testing the flag. Only ENABLED tells an absent component from a present one.
	engine := &CacheEngine{name: "demo", namespace: "default"}
	script, err := engine.generateRuntimeSh(&common.CacheRuntimeConfig{
		Master: &common.CacheRuntimeComponentConfig{Enabled: true, Name: "demo-master"},
		Client: nil,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// The master is present but carries no tiered store either, so the two must export the
	// same variables. Anything the master gains and the absent client does not would be
	// unset for a consumer, which is what this guards against.
	suffixes := func(prefix string) []string {
		var found []string
		for _, match := range regexp.MustCompile(`(?m)^export ([A-Z0-9_]+)=`).FindAllStringSubmatch(script, -1) {
			if strings.HasPrefix(match[1], prefix) {
				found = append(found, strings.TrimPrefix(match[1], prefix))
			}
		}
		sort.Strings(found)
		return found
	}
	master, client := suffixes("MASTER_"), suffixes("CLIENT_")
	if len(master) == 0 {
		t.Fatal("expected the master to export something")
	}
	if !reflect.DeepEqual(master, client) {
		t.Fatalf("an absent component must export the same variables as a present one.\n"+
			"master: %v\nclient: %v\nCheck writeComponentSh in cm.go.", master, client)
	}

	values := sourceRuntimeShNounset(t, script,
		"CLIENT_ENABLED", "CLIENT_NAME", "CLIENT_REPLICAS", "CLIENT_SERVICE_NAME",
		"CLIENT_OPTIONS", "CLIENT_TIEREDSTORE_LEVELS_COUNT")
	expected := []string{"false", "", "0", "", "{}", "0"}
	for i, want := range expected {
		if values[i] != want {
			t.Fatalf("value %d: expected %q, got %q", i, want, values[i])
		}
	}
}

// sourceRuntimeShNounset sources the script with "set -u" in effect, so that reading an
// unexported variable fails the test rather than yielding an empty string.
func sourceRuntimeShNounset(t *testing.T, script string, keys ...string) []string {
	t.Helper()
	shell, err := exec.LookPath("sh")
	if err != nil {
		t.Skipf("sh is not available: %v", err)
	}

	path := filepath.Join(t.TempDir(), "runtime.sh")
	if err := os.WriteFile(path, []byte(script), 0o600); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}

	const sep = "@@FLUID@@"
	const prog = `set -u; . "$1"; shift; for k in "$@"; do eval "printf '%s@@FLUID@@' \"\$$k\""; done`
	args := append([]string{"-c", prog, "sh", path}, keys...)
	out, err := exec.Command(shell, args...).Output()
	if err != nil {
		t.Fatalf("sourcing under \"set -u\" failed, so a variable is unset: %v\nscript:\n%s", err, script)
	}

	values := strings.Split(string(out), sep)
	values = values[:len(values)-1]
	if len(values) != len(keys) {
		t.Fatalf("expected %d values, got %d (%q)", len(keys), len(values), values)
	}
	return values
}

func sourceRuntimeSh(t *testing.T, script string, keys ...string) []string {
	t.Helper()
	shell, err := exec.LookPath("sh")
	if err != nil {
		t.Skipf("sh is not available: %v", err)
	}

	path := filepath.Join(t.TempDir(), "runtime.sh")
	if err := os.WriteFile(path, []byte(script), 0o600); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}

	const sep = "@@FLUID@@"
	const prog = `. "$1"; shift; for k in "$@"; do eval "printf '%s@@FLUID@@' \"\$$k\""; done`
	args := append([]string{"-c", prog, "sh", path}, keys...)
	out, err := exec.Command(shell, args...).Output()
	if err != nil {
		t.Fatalf("failed to source generated script: %v\nscript:\n%s", err, script)
	}

	values := strings.Split(string(out), sep)
	// The script prints a trailing separator after the last value.
	values = values[:len(values)-1]
	if len(values) != len(keys) {
		t.Fatalf("expected %d values, got %d (%q)", len(keys), len(values), values)
	}
	return values
}

func TestGenerateRuntimeConfigDataIncludesRuntimeSh(t *testing.T) {
	runtimeObj := newClientlessRuntimeForConfigMapTest()
	engine := newEngineForConfigMapTest(t, runtimeObj, newCacheRuntimeClassForConfigMapTest(), newDatasetForConfigMapTest())

	data, err := engine.generateRuntimeConfigData(context.Background(), runtimeObj)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// runtime.json stays exactly what it was: existing components keep parsing it.
	raw, ok := data[engine.getRuntimeConfigFileName()]
	if !ok {
		t.Fatalf("expected key %q, got keys %v", engine.getRuntimeConfigFileName(), data)
	}
	config := common.CacheRuntimeConfig{}
	if err := json.Unmarshal([]byte(raw), &config); err != nil {
		t.Fatalf("expected runtime.json to stay valid JSON, got %v", err)
	}

	script, ok := data[engine.getRuntimeShFileName()]
	if !ok {
		t.Fatalf("expected key %q, got keys %v", engine.getRuntimeShFileName(), data)
	}
	if !strings.HasPrefix(script, "#!/bin/sh\n") {
		t.Fatalf("expected a shebang, got %q", script[:min(len(script), 40)])
	}

	// The script must agree with the JSON it was generated from, field by field.
	values := sourceRuntimeSh(t, script,
		"RUNTIME_TARGETPATH", "RUNTIME_ACCESSMODES_COUNT", "RUNTIME_ACCESSMODE_0",
		"MOUNTS_COUNT", "MOUNT_0_NAME", "MOUNT_0_PATH", "MOUNT_0_MOUNTPOINT", "MOUNT_0_READONLY", "MOUNT_0_SHARED",
		"MASTER_ENABLED", "MASTER_NAME", "MASTER_REPLICAS",
		"WORKER_ENABLED", "WORKER_NAME", "WORKER_REPLICAS",
		"CLIENT_ENABLED")
	expected := []string{
		config.TargetPath, "1", "ReadOnlyMany",
		"1", config.Mounts[0].Name, config.Mounts[0].Path, config.Mounts[0].MountPoint, "true", "true",
		"true", config.Master.Name, "1",
		"true", config.Worker.Name, "3",
		// The client is disabled, so it exports only its flag. Consumers can test
		// "$CLIENT_ENABLED" without having to check whether the variable is set at all.
		"false",
	}
	for i, want := range expected {
		if values[i] != want {
			t.Fatalf("value %d: expected %q, got %q", i, want, values[i])
		}
	}
}

func TestGenerateRuntimeShQuotesHostileValues(t *testing.T) {
	// A mount option is user input copied into the script verbatim. Sourcing it must not
	// let the value break out of its quotes or run anything.
	hostile := `it's a "value" $(exit 7) ` + "`exit 7`" + " with\nnewline"
	dataset := newDatasetForConfigMapTest()
	dataset.Spec.Mounts[0].Options = map[string]string{"hostile": hostile}
	dataset.Spec.Mounts[0].Name = "it's-a-mount"

	runtimeObj := newClientlessRuntimeForConfigMapTest()
	engine := newEngineForConfigMapTest(t, runtimeObj, newCacheRuntimeClassForConfigMapTest(), dataset)

	data, err := engine.generateRuntimeConfigData(context.Background(), runtimeObj)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	values := sourceRuntimeSh(t, data[engine.getRuntimeShFileName()], "MOUNT_0_OPTIONS", "MOUNT_0_NAME")

	wantOptions, err := json.Marshal(map[string]string{"hostile": hostile})
	if err != nil {
		t.Fatal(err)
	}
	if values[0] != string(wantOptions) {
		t.Fatalf("expected options %q, got %q", wantOptions, values[0])
	}
	if values[1] != "it's-a-mount" {
		t.Fatalf("expected the mount name to survive quoting, got %q", values[1])
	}
}

func TestGenerateRuntimeShIsDeterministic(t *testing.T) {
	// generateRuntimeConfigData walks Go maps, whose iteration order is randomized. If any
	// of that order leaked into the script, syncRuntimeValueConfigMap would see the data
	// change on every reconcile and update the ConfigMap forever.
	dataset := newDatasetForConfigMapTest()
	dataset.Spec.Mounts[0].Options = map[string]string{"a": "1", "b": "2", "c": "3", "d": "4", "e": "5"}

	runtimeObj := newClientlessRuntimeForConfigMapTest()
	runtimeObj.Spec.Options = map[string]string{"x": "1", "y": "2", "z": "3"}
	runtimeObj.Spec.Master.Options = map[string]string{"m1": "1", "m2": "2", "m3": "3"}

	engine := newEngineForConfigMapTest(t, runtimeObj, newCacheRuntimeClassForConfigMapTest(), dataset)

	first, err := engine.generateRuntimeConfigData(context.Background(), runtimeObj)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	for i := 0; i < 20; i++ {
		next, err := engine.generateRuntimeConfigData(context.Background(), runtimeObj)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if next[engine.getRuntimeShFileName()] != first[engine.getRuntimeShFileName()] {
			t.Fatalf("runtime.sh is not stable across calls:\nfirst:\n%s\nlater:\n%s",
				first[engine.getRuntimeShFileName()], next[engine.getRuntimeShFileName()])
		}
	}
}

func TestGenerateRuntimeShWithoutAnyEnabledComponent(t *testing.T) {
	// Every component disabled: the script still sources cleanly and reports all three as
	// off, instead of leaving the variables unset.
	runtimeObj := newCacheRuntimeForConfigMapTest()
	engine := newEngineForConfigMapTest(t, runtimeObj, newCacheRuntimeClassForConfigMapTest(), newDatasetForConfigMapTest())

	data, err := engine.generateRuntimeConfigData(context.Background(), runtimeObj)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	values := sourceRuntimeSh(t, data[engine.getRuntimeShFileName()], "MASTER_ENABLED", "WORKER_ENABLED", "CLIENT_ENABLED")
	for i, got := range values {
		if got != "false" {
			t.Fatalf("value %d: expected %q, got %q", i, "false", got)
		}
	}
}

func TestGenerateRuntimeShExportsTieredStoreLevels(t *testing.T) {
	quota := resource.MustParse("10Gi")
	runtimeObj := newClientlessRuntimeForConfigMapTest()
	runtimeObj.Spec.Worker.TieredStore = datav1alpha1.RuntimeTieredStore{
		Levels: []datav1alpha1.RuntimeTieredStoreLevel{{
			ProcessMemory: &datav1alpha1.ProcessMemoryMediumSource{Quota: quota},
			High:          "0.95",
			Low:           "0.7",
		}},
	}
	engine := newEngineForConfigMapTest(t, runtimeObj, newCacheRuntimeClassForConfigMapTest(), newDatasetForConfigMapTest())

	data, err := engine.generateRuntimeConfigData(context.Background(), runtimeObj)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	values := sourceRuntimeSh(t, data[engine.getRuntimeShFileName()],
		"WORKER_TIEREDSTORE_LEVELS_COUNT",
		"WORKER_TIEREDSTORE_LEVEL_0_MEDIUMTYPE",
		"WORKER_TIEREDSTORE_LEVEL_0_MOUNTPATHS_COUNT",
		"WORKER_TIEREDSTORE_LEVEL_0_MOUNTPATH_0",
		"WORKER_TIEREDSTORE_LEVEL_0_QUOTAS_COUNT",
		"WORKER_TIEREDSTORE_LEVEL_0_QUOTA_0",
		"WORKER_TIEREDSTORE_LEVEL_0_HIGH",
		"WORKER_TIEREDSTORE_LEVEL_0_LOW")
	expected := []string{"1", string(common.Memory), "1", GetMemoryTieredStoreMountPath(0), "1", quota.String(), "0.95", "0.7"}
	for i, want := range expected {
		if values[i] != want {
			t.Fatalf("value %d: expected %q, got %q", i, want, values[i])
		}
	}
}

// --- drift guard -------------------------------------------------------------------
//
// generateRuntimeSh maps the runtime config onto shell variables by hand, so that the
// names are chosen rather than derived. The cost of that choice is that a field added to
// the config later is simply not exported, silently: the ConfigMap still generates, every
// existing test still passes, and the client only finds out at runtime that the value it
// needs is not in runtime.sh.
//
// The table below is the other half of that hand-written mapping, and the test walks the
// config types with reflection to check the two agree. Adding a field to the config
// without exporting it fails here, and names the field.

type shExport struct {
	// variable is the shell variable the field must end up in, for the fixture below.
	variable string
	// value is what sourcing runtime.sh must leave in that variable.
	value string
}

// runtimeShExports covers every field of every type reachable from CacheRuntimeConfig,
// keyed by "<type>.<field>". Component fields are checked through the master; the worker
// is checked separately, by comparing the two prefixes' variable sets.
var runtimeShExports = map[string]shExport{
	"CacheRuntimeConfig.TargetPath":  {"RUNTIME_TARGETPATH", "/sentinel/target-path"},
	"CacheRuntimeConfig.AccessModes": {"RUNTIME_ACCESSMODES_COUNT", "2"},
	"CacheRuntimeConfig.Mounts":      {"MOUNTS_COUNT", "2"},
	"CacheRuntimeConfig.Master":      {"MASTER_ENABLED", "true"},
	"CacheRuntimeConfig.Worker":      {"WORKER_ENABLED", "true"},
	// A nil component still reports itself, so consumers can test the flag without
	// having to check whether the variable is set at all.
	"CacheRuntimeConfig.Client": {"CLIENT_ENABLED", "false"},

	"MountConfig.Name":           {"MOUNT_0_NAME", "sentinel-mount-name"},
	"MountConfig.Path":           {"MOUNT_0_PATH", "/sentinel/mount-path"},
	"MountConfig.MountPoint":     {"MOUNT_0_MOUNTPOINT", "sentinel:///mount-point"},
	"MountConfig.ReadOnly":       {"MOUNT_0_READONLY", "true"},
	"MountConfig.Shared":         {"MOUNT_0_SHARED", "false"},
	"MountConfig.Options":        {"MOUNT_0_OPTIONS", `{"sentinel-option":"sentinel-option-value"}`},
	"MountConfig.EncryptOptions": {"MOUNT_0_ENCRYPTOPTIONS", `{"sentinel-encrypt":"/sentinel/secret"}`},

	"CacheRuntimeComponentConfig.Enabled":           {"MASTER_ENABLED", "true"},
	"CacheRuntimeComponentConfig.Name":              {"MASTER_NAME", "sentinel-master-name"},
	"CacheRuntimeComponentConfig.Replicas":          {"MASTER_REPLICAS", "9001"},
	"CacheRuntimeComponentConfig.Options":           {"MASTER_OPTIONS", `{"sentinel-master-option":"sentinel-master-value"}`},
	"CacheRuntimeComponentConfig.Service":           {"MASTER_SERVICE_NAME", "sentinel-master-service"},
	"CacheRuntimeComponentConfig.TieredStoreLevels": {"MASTER_TIEREDSTORE_LEVELS_COUNT", "1"},

	"CacheRuntimeComponentServiceConfig.Name": {"MASTER_SERVICE_NAME", "sentinel-master-service"},

	"TieredStoreLevelConfig.MediumType": {"MASTER_TIEREDSTORE_LEVEL_0_MEDIUMTYPE", string(common.Memory)},
	"TieredStoreLevelConfig.MountPaths": {"MASTER_TIEREDSTORE_LEVEL_0_MOUNTPATHS_COUNT", "2"},
	"TieredStoreLevelConfig.Quotas":     {"MASTER_TIEREDSTORE_LEVEL_0_QUOTAS_COUNT", "2"},
	"TieredStoreLevelConfig.High":       {"MASTER_TIEREDSTORE_LEVEL_0_HIGH", "0.91"},
	"TieredStoreLevelConfig.Low":        {"MASTER_TIEREDSTORE_LEVEL_0_LOW", "0.71"},
}

// fullyPopulatedRuntimeConfig sets every field named in runtimeShExports to a value that
// is distinguishable from every other field's, so that a mapping that reaches the wrong
// field is caught as well as one that reaches no field at all.
func fullyPopulatedRuntimeConfig() *common.CacheRuntimeConfig {
	component := func(prefix string) *common.CacheRuntimeComponentConfig {
		return &common.CacheRuntimeComponentConfig{
			Enabled:  true,
			Name:     "sentinel-" + prefix + "-name",
			Replicas: 9001,
			Options:  map[string]string{"sentinel-" + prefix + "-option": "sentinel-" + prefix + "-value"},
			Service:  common.CacheRuntimeComponentServiceConfig{Name: "sentinel-" + prefix + "-service"},
			TieredStoreLevels: []common.TieredStoreLevelConfig{{
				MediumType: common.Memory,
				MountPaths: []string{"/sentinel/tier-a", "/sentinel/tier-b"},
				Quotas:     []string{"11Gi", "22Gi"},
				High:       "0.91",
				Low:        "0.71",
			}},
		}
	}

	return &common.CacheRuntimeConfig{
		TargetPath:  "/sentinel/target-path",
		AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadOnlyMany, corev1.ReadWriteMany},
		Mounts: []common.MountConfig{
			{
				Name:           "sentinel-mount-name",
				Path:           "/sentinel/mount-path",
				MountPoint:     "sentinel:///mount-point",
				ReadOnly:       true,
				Shared:         false,
				Options:        map[string]string{"sentinel-option": "sentinel-option-value"},
				EncryptOptions: map[string]string{"sentinel-encrypt": "/sentinel/secret"},
			},
			// A second mount, so that MOUNTS_COUNT and the per-element prefixes are
			// exercised rather than assumed.
			{Name: "sentinel-second-mount"},
		},
		Master: component("master"),
		Worker: component("worker"),
		// Left nil on purpose: the client is the component a FUSE-less runtime switches off.
		Client: nil,
	}
}

// configFieldPaths walks the runtime config types and returns every field as
// "<type>.<field>". Struct types are walked once, however many times they are reached.
func configFieldPaths(t *testing.T, typ reflect.Type, seen map[reflect.Type]bool, paths map[string]bool) {
	t.Helper()

	for typ.Kind() == reflect.Ptr || typ.Kind() == reflect.Slice || typ.Kind() == reflect.Map {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct || seen[typ] {
		return
	}
	seen[typ] = true

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" {
			// unexported, never serialized, so never exported to the shell either
			continue
		}
		paths[typ.Name()+"."+field.Name] = true
		configFieldPaths(t, field.Type, seen, paths)
	}
}

func TestRuntimeShCoversEveryConfigField(t *testing.T) {
	paths := map[string]bool{}
	configFieldPaths(t, reflect.TypeOf(common.CacheRuntimeConfig{}), map[reflect.Type]bool{}, paths)

	for path := range paths {
		if _, ok := runtimeShExports[path]; !ok {
			t.Errorf("%s exists in the runtime config but is not exported to runtime.sh.\n"+
				"Add an export for it in generateRuntimeSh (cm.go), then record the variable it "+
				"maps to in runtimeShExports and give the field a value in fullyPopulatedRuntimeConfig.", path)
		}
	}
	for path := range runtimeShExports {
		if !paths[path] {
			t.Errorf("runtimeShExports lists %s, which no longer exists in the runtime config.\n"+
				"Drop it here, and drop its export from generateRuntimeSh (cm.go).", path)
		}
	}
	if t.Failed() {
		return
	}

	engine := &CacheEngine{name: "demo", namespace: "default"}
	script, err := engine.generateRuntimeSh(fullyPopulatedRuntimeConfig())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Sourcing the script is what the consumer does, so the values are read back the same
	// way rather than pattern-matched out of the text.
	fields := make([]string, 0, len(runtimeShExports))
	for path := range runtimeShExports {
		fields = append(fields, path)
	}
	sort.Strings(fields)

	variables := make([]string, 0, len(fields))
	for _, path := range fields {
		variables = append(variables, runtimeShExports[path].variable)
	}
	values := sourceRuntimeSh(t, script, variables...)

	for i, path := range fields {
		if want := runtimeShExports[path].value; values[i] != want {
			t.Errorf("%s: expected %s=%q, got %q.\n"+
				"Either generateRuntimeSh (cm.go) no longer exports this field as %s, or it exports "+
				"a different field's value into it.", path, variables[i], want, values[i], variables[i])
		}
	}
}

func TestRuntimeShExportsEveryComponentAlike(t *testing.T) {
	// The component mapping above is checked through the master only. The worker is
	// populated identically, so anything the master exports the worker must export too:
	// this catches a hand-written export that was added for one component and forgotten
	// for the other.
	engine := &CacheEngine{name: "demo", namespace: "default"}
	script, err := engine.generateRuntimeSh(fullyPopulatedRuntimeConfig())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	suffixes := func(prefix string) []string {
		var found []string
		for _, match := range regexp.MustCompile(`(?m)^export ([A-Z0-9_]+)=`).FindAllStringSubmatch(script, -1) {
			if strings.HasPrefix(match[1], prefix) {
				found = append(found, strings.TrimPrefix(match[1], prefix))
			}
		}
		sort.Strings(found)
		return found
	}

	master, worker := suffixes("MASTER_"), suffixes("WORKER_")
	if len(master) == 0 {
		t.Fatal("expected the master to export something")
	}
	if !reflect.DeepEqual(master, worker) {
		t.Fatalf("the master and the worker are populated identically but export different variables.\n"+
			"master: %v\nworker: %v\nCheck writeComponentSh in cm.go.", master, worker)
	}
}
