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

package thin

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

// chart-injected fuse pod template entries, see charts/thin/templates/fuse/daemonset.yaml. They
// exist alongside the ones derived from .Values.fuse, so syncing must not drop them.
var (
	chartFuseVolumeNames = []string{"thin-fuse-mount", "thin-conf"}
	chartFuseEnvNames    = []string{"FLUID_RUNTIME_TYPE", "FLUID_RUNTIME_NS", "FLUID_RUNTIME_NAME"}
)

// syncRuntimeFixture is a ThinRuntime that has finished setup: the values ConfigMap and the fuse
// DaemonSet both hold the state rendered from the initial spec.
type syncRuntimeFixture struct {
	engine       *ThinEngine
	runtime      *datav1alpha1.ThinRuntime
	initialValue *ThinValue
}

// newSyncRuntimeFixture renders the initial spec the way setupMasterInternal does, then seeds the
// resulting values ConfigMap and fuse DaemonSet, so that a test only has to edit the ThinRuntime.
func newSyncRuntimeFixture(t *testing.T, setup func(*datav1alpha1.ThinRuntime, *datav1alpha1.ThinRuntimeProfile)) *syncRuntimeFixture {
	t.Helper()

	dataset, runtime, profile := mockFluidObjectsForTests(types.NamespacedName{Name: "test-dataset", Namespace: "fluid"})
	// A pvc mountPoint would need a bound PV, which is out of scope here.
	dataset.Spec.Mounts = []datav1alpha1.Mount{{MountPoint: "nfs://192.168.0.1/data", Name: "data"}}
	runtime.Spec.Fuse = datav1alpha1.ThinFuseSpec{
		Image:           "fluidcloudnative/nfs",
		ImageTag:        "v0.1",
		ImagePullPolicy: string(corev1.PullIfNotPresent),
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1")},
		},
		Env: []corev1.EnvVar{{Name: "FROM_RUNTIME", Value: "initial"}},
	}
	if setup != nil {
		setup(runtime, profile)
	}

	engine := mockThinEngineForTests(dataset, runtime, profile)
	engine.Client = fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme, dataset, runtime, profile)

	initialValue, err := engine.transform(runtime, profile)
	if err != nil {
		t.Fatalf("failed to render the initial value: %v", err)
	}

	if err := engine.Client.Create(context.TODO(), newValuesConfigMapForTests(t, engine, initialValue)); err != nil {
		t.Fatalf("failed to create the values configmap: %v", err)
	}
	if err := engine.Client.Create(context.TODO(), newFuseDaemonSetForTests(t, engine, initialValue)); err != nil {
		t.Fatalf("failed to create the fuse daemonset: %v", err)
	}

	return &syncRuntimeFixture{engine: engine, runtime: runtime, initialValue: initialValue}
}

// editRuntime applies a spec change the way an operator would, after the runtime is ready.
func (f *syncRuntimeFixture) editRuntime(t *testing.T, edit func(*datav1alpha1.ThinRuntime)) {
	t.Helper()

	runtime := &datav1alpha1.ThinRuntime{}
	if err := f.engine.Client.Get(context.TODO(),
		types.NamespacedName{Name: f.engine.name, Namespace: f.engine.namespace}, runtime); err != nil {
		t.Fatalf("failed to get the runtime: %v", err)
	}
	edit(runtime)
	if err := f.engine.Client.Update(context.TODO(), runtime); err != nil {
		t.Fatalf("failed to update the runtime: %v", err)
	}
}

func (f *syncRuntimeFixture) fuseDaemonSet(t *testing.T) *appsv1.DaemonSet {
	t.Helper()

	ds := &appsv1.DaemonSet{}
	if err := f.engine.Client.Get(context.TODO(),
		types.NamespacedName{Name: f.engine.getFuseName(), Namespace: f.engine.namespace}, ds); err != nil {
		t.Fatalf("failed to get the fuse daemonset: %v", err)
	}
	return ds
}

func (f *syncRuntimeFixture) syncedValue(t *testing.T) *ThinValue {
	t.Helper()

	value, err := f.engine.GetValueFromConfigmap()
	if err != nil {
		t.Fatalf("failed to read the values configmap: %v", err)
	}
	if value == nil {
		t.Fatal("the values configmap is unexpectedly missing")
	}
	return value
}

func (f *syncRuntimeFixture) fuseContainer(t *testing.T) corev1.Container {
	t.Helper()

	ds := f.fuseDaemonSet(t)
	idx := utils.GetContainerIndex(ds.Spec.Template.Spec.Containers, fuseContainerName)
	if idx < 0 {
		t.Fatalf("container %s not found in the fuse daemonset", fuseContainerName)
	}
	return ds.Spec.Template.Spec.Containers[idx]
}

func newValuesConfigMapForTests(t *testing.T, engine *ThinEngine, value *ThinValue) *corev1.ConfigMap {
	t.Helper()

	data, err := yaml.Marshal(value)
	if err != nil {
		t.Fatalf("failed to marshal the value: %v", err)
	}

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      engine.getHelmValuesConfigMapName(),
			Namespace: engine.namespace,
		},
		Data: map[string]string{"data": string(data)},
	}
}

// newFuseDaemonSetForTests mimics what helm renders from the thin chart for the given value,
// including the entries the chart adds on top of .Values.fuse.
func newFuseDaemonSetForTests(t *testing.T, engine *ThinEngine, value *ThinValue) *appsv1.DaemonSet {
	t.Helper()

	resources, err := utils.TransformInternalResourcesToCoreV1Resources(value.Fuse.Resources)
	if err != nil {
		t.Fatalf("failed to transform the resources: %v", err)
	}

	chartEnvs := []corev1.EnvVar{
		{Name: "FLUID_RUNTIME_TYPE", Value: "thin"},
		{Name: "FLUID_RUNTIME_NS", Value: value.RuntimeIdentity.Namespace},
		{Name: "FLUID_RUNTIME_NAME", Value: value.RuntimeIdentity.Name},
	}
	chartVolumes := []corev1.Volume{
		{Name: "thin-fuse-mount", VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{Path: "/runtime-mnt/thin/fluid/test-dataset"}}},
		{Name: "thin-conf", VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: engine.getFuseConfigMapName()}}}},
	}
	chartVolumeMounts := []corev1.VolumeMount{
		{Name: "thin-fuse-mount", MountPath: "/runtime-mnt/thin/fluid/test-dataset"},
		{Name: "thin-conf", MountPath: "/etc/fluid/config", ReadOnly: true},
	}

	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      engine.getFuseName(),
			Namespace: engine.namespace,
		},
		Spec: appsv1.DaemonSetSpec{
			UpdateStrategy: appsv1.DaemonSetUpdateStrategy{Type: appsv1.OnDeleteDaemonSetStrategyType},
			Selector:       &metav1.LabelSelector{MatchLabels: map[string]string{"role": "thin-fuse"}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: utils.UnionMapsWithOverride(
						map[string]string{"role": "thin-fuse", "app": "thin"}, value.Fuse.Labels),
					Annotations: utils.UnionMapsWithOverride(
						map[string]string{"sidecar.istio.io/inject": "false"}, value.Fuse.Annotations),
				},
				Spec: corev1.PodSpec{
					NodeSelector: value.Fuse.NodeSelector,
					Volumes:      append(chartVolumes, value.Fuse.Volumes...),
					Containers: []corev1.Container{{
						Name:            fuseContainerName,
						Image:           composeImage(value.Fuse.Image, value.Fuse.ImageTag),
						ImagePullPolicy: corev1.PullPolicy(value.Fuse.ImagePullPolicy),
						Resources:       resources,
						Env:             append(chartEnvs, value.Fuse.Envs...),
						VolumeMounts:    append(chartVolumeMounts, value.Fuse.VolumeMounts...),
						Lifecycle:       value.Fuse.Lifecycle,
					}},
				},
			},
		},
	}
}

func TestSyncRuntimeConvergesFuseTemplate(t *testing.T) {
	testCases := []struct {
		name   string
		edit   func(*datav1alpha1.ThinRuntime)
		verify func(*testing.T, *syncRuntimeFixture)
	}{
		{
			name: "resources",
			edit: func(runtime *datav1alpha1.ThinRuntime) {
				runtime.Spec.Fuse.Resources.Requests = corev1.ResourceList{
					corev1.ResourceCPU: resource.MustParse("4"),
				}
				runtime.Spec.Fuse.Resources.Limits = corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("2Gi"),
				}
			},
			verify: func(t *testing.T, f *syncRuntimeFixture) {
				container := f.fuseContainer(t)
				if got := container.Resources.Requests[corev1.ResourceCPU]; got.String() != "4" {
					t.Errorf("daemonset cpu request = %s, want 4", got.String())
				}
				if got := container.Resources.Limits[corev1.ResourceMemory]; got.String() != "2Gi" {
					t.Errorf("daemonset memory limit = %s, want 2Gi", got.String())
				}

				value := f.syncedValue(t)
				if got := value.Fuse.Resources.Requests[corev1.ResourceCPU]; got != "4" {
					t.Errorf("values configmap cpu request = %q, want \"4\"", got)
				}
				if got := value.Fuse.Resources.Limits[corev1.ResourceMemory]; got != "2Gi" {
					t.Errorf("values configmap memory limit = %q, want \"2Gi\"", got)
				}
			},
		},
		{
			name: "image and tag",
			edit: func(runtime *datav1alpha1.ThinRuntime) {
				runtime.Spec.Fuse.Image = "fluidcloudnative/nfs"
				runtime.Spec.Fuse.ImageTag = "v0.2"
			},
			verify: func(t *testing.T, f *syncRuntimeFixture) {
				if got, want := f.fuseContainer(t).Image, "fluidcloudnative/nfs:v0.2"; got != want {
					t.Errorf("daemonset image = %q, want %q", got, want)
				}

				value := f.syncedValue(t)
				if got, want := value.Fuse.ImageTag, "v0.2"; got != want {
					t.Errorf("values configmap image tag = %q, want %q", got, want)
				}
			},
		},
		{
			name: "image pull policy",
			edit: func(runtime *datav1alpha1.ThinRuntime) {
				runtime.Spec.Fuse.ImagePullPolicy = string(corev1.PullAlways)
			},
			verify: func(t *testing.T, f *syncRuntimeFixture) {
				if got := f.fuseContainer(t).ImagePullPolicy; got != corev1.PullAlways {
					t.Errorf("daemonset image pull policy = %q, want %q", got, corev1.PullAlways)
				}
				if got := f.syncedValue(t).Fuse.ImagePullPolicy; got != string(corev1.PullAlways) {
					t.Errorf("values configmap image pull policy = %q, want %q", got, corev1.PullAlways)
				}
			},
		},
		{
			name: "lifecycle",
			edit: func(runtime *datav1alpha1.ThinRuntime) {
				runtime.Spec.Fuse.Lifecycle = &corev1.Lifecycle{
					PreStop: &corev1.LifecycleHandler{
						Exec: &corev1.ExecAction{Command: []string{"/bin/sh", "-c", "custom-teardown"}},
					},
				}
			},
			verify: func(t *testing.T, f *syncRuntimeFixture) {
				lifecycle := f.fuseContainer(t).Lifecycle
				if lifecycle == nil || lifecycle.PreStop == nil || lifecycle.PreStop.Exec == nil {
					t.Fatalf("daemonset preStop exec handler is missing, got %v", lifecycle)
				}
				if got, want := lifecycle.PreStop.Exec.Command[2], "custom-teardown"; got != want {
					t.Errorf("daemonset preStop command = %q, want %q", got, want)
				}

				value := f.syncedValue(t)
				if value.Fuse.Lifecycle == nil || value.Fuse.Lifecycle.PreStop == nil ||
					value.Fuse.Lifecycle.PreStop.Exec == nil {
					t.Fatalf("values configmap preStop exec handler is missing, got %v", value.Fuse.Lifecycle)
				}
				if got, want := value.Fuse.Lifecycle.PreStop.Exec.Command[2], "custom-teardown"; got != want {
					t.Errorf("values configmap preStop command = %q, want %q", got, want)
				}
			},
		},
		{
			name: "env",
			edit: func(runtime *datav1alpha1.ThinRuntime) {
				runtime.Spec.Fuse.Env = []corev1.EnvVar{
					{Name: "FROM_RUNTIME", Value: "updated"},
					{Name: "EXTRA", Value: "added"},
				}
			},
			verify: func(t *testing.T, f *syncRuntimeFixture) {
				envs := envsByName(f.fuseContainer(t).Env)

				// The chart's own env variables must survive the sync.
				for _, name := range chartFuseEnvNames {
					if _, ok := envs[name]; !ok {
						t.Errorf("chart env variable %s was dropped from the daemonset", name)
					}
				}
				if got, want := envs["FROM_RUNTIME"], "updated"; got != want {
					t.Errorf("daemonset FROM_RUNTIME = %q, want %q", got, want)
				}
				if got, want := envs["EXTRA"], "added"; got != want {
					t.Errorf("daemonset EXTRA = %q, want %q", got, want)
				}
				// The fuse mount point env variable is derived by transformFuse, not by the chart.
				if _, ok := envs[common.ThinFusePointEnvKey]; !ok {
					t.Errorf("env variable %s was dropped from the daemonset", common.ThinFusePointEnvKey)
				}

				syncedEnvs := envsByName(f.syncedValue(t).Fuse.Envs)
				if got, want := syncedEnvs["EXTRA"], "added"; got != want {
					t.Errorf("values configmap EXTRA = %q, want %q", got, want)
				}
			},
		},
		{
			name: "mount options render into the env",
			edit: func(runtime *datav1alpha1.ThinRuntime) {
				runtime.Spec.Fuse.Options = map[string]string{"vers": "4.1", "nolock": ""}
			},
			verify: func(t *testing.T, f *syncRuntimeFixture) {
				got := envsByName(f.fuseContainer(t).Env)[common.ThinFuseOptionEnvKey]
				// Sorted, so the rendered string is stable across reconciliations.
				if want := "nolock,ro,vers=4.1"; got != want {
					t.Errorf("daemonset %s = %q, want %q", common.ThinFuseOptionEnvKey, got, want)
				}
			},
		},
		{
			name: "volumes and volume mounts",
			edit: func(runtime *datav1alpha1.ThinRuntime) {
				runtime.Spec.Volumes = []corev1.Volume{{
					Name:         "extra",
					VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
				}}
				runtime.Spec.Fuse.VolumeMounts = []corev1.VolumeMount{{Name: "extra", MountPath: "/extra"}}
			},
			verify: func(t *testing.T, f *syncRuntimeFixture) {
				ds := f.fuseDaemonSet(t)

				volumes := map[string]bool{}
				for _, v := range ds.Spec.Template.Spec.Volumes {
					volumes[v.Name] = true
				}
				// The chart's own volumes must survive the sync.
				for _, name := range chartFuseVolumeNames {
					if !volumes[name] {
						t.Errorf("chart volume %s was dropped from the daemonset", name)
					}
				}
				if !volumes["extra"] {
					t.Errorf("volume extra was not added to the daemonset, got %v", volumes)
				}

				mounts := map[string]string{}
				for _, m := range f.fuseContainer(t).VolumeMounts {
					mounts[m.Name] = m.MountPath
				}
				for _, name := range chartFuseVolumeNames {
					if _, ok := mounts[name]; !ok {
						t.Errorf("chart volume mount %s was dropped from the daemonset", name)
					}
				}
				if got, want := mounts["extra"], "/extra"; got != want {
					t.Errorf("volume mount extra = %q, want %q", got, want)
				}
			},
		},
		{
			name: "pod labels and annotations",
			edit: func(runtime *datav1alpha1.ThinRuntime) {
				runtime.Spec.Fuse.PodMetadata = datav1alpha1.PodMetadata{
					Labels:      map[string]string{"team": "storage"},
					Annotations: map[string]string{"owner": "platform"},
				}
			},
			verify: func(t *testing.T, f *syncRuntimeFixture) {
				template := f.fuseDaemonSet(t).Spec.Template.ObjectMeta

				if got, want := template.Labels["team"], "storage"; got != want {
					t.Errorf("daemonset pod label team = %q, want %q", got, want)
				}
				// The chart's own labels must survive the sync.
				if got, want := template.Labels["role"], "thin-fuse"; got != want {
					t.Errorf("chart pod label role = %q, want %q", got, want)
				}
				if got, want := template.Annotations["owner"], "platform"; got != want {
					t.Errorf("daemonset pod annotation owner = %q, want %q", got, want)
				}
				if got, want := template.Annotations["sidecar.istio.io/inject"], "false"; got != want {
					t.Errorf("chart pod annotation sidecar.istio.io/inject = %q, want %q", got, want)
				}
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			f := newSyncRuntimeFixture(t, nil)
			f.editRuntime(t, testCase.edit)

			changed, err := f.engine.SyncRuntime(cruntime.ReconcileRequestContext{})
			if err != nil {
				t.Fatalf("SyncRuntime returned an error: %v", err)
			}
			if !changed {
				t.Fatal("SyncRuntime reported no change, want the fuse template to be updated")
			}

			testCase.verify(t, f)

			// A second sync must be a no-op, otherwise every reconciliation would churn the
			// daemonset and spam events.
			changed, err = f.engine.SyncRuntime(cruntime.ReconcileRequestContext{})
			if err != nil {
				t.Fatalf("the second SyncRuntime returned an error: %v", err)
			}
			if changed {
				t.Error("the second SyncRuntime reported a change, want the sync to be idempotent")
			}
		})
	}
}

// TestSyncRuntimeIsANoOpWithoutSpecChanges guards the idempotence of the freshly rendered state:
// syncing a runtime nobody touched must not update anything.
func TestSyncRuntimeIsANoOpWithoutSpecChanges(t *testing.T) {
	f := newSyncRuntimeFixture(t, nil)
	before := f.fuseDaemonSet(t)

	changed, err := f.engine.SyncRuntime(cruntime.ReconcileRequestContext{})
	if err != nil {
		t.Fatalf("SyncRuntime returned an error: %v", err)
	}
	if changed {
		t.Error("SyncRuntime reported a change, want no change")
	}

	if after := f.fuseDaemonSet(t); after.ResourceVersion != before.ResourceVersion {
		t.Errorf("the fuse daemonset was updated, resourceVersion %s -> %s",
			before.ResourceVersion, after.ResourceVersion)
	}
}

// TestSyncRuntimeCoercesUnsafeUpdateStrategy makes sure the template is never pushed into a
// RollingUpdate daemonset, which would restart the fuse pods and break the applications using them.
func TestSyncRuntimeCoercesUnsafeUpdateStrategy(t *testing.T) {
	f := newSyncRuntimeFixture(t, nil)

	ds := f.fuseDaemonSet(t)
	ds.Spec.UpdateStrategy = appsv1.DaemonSetUpdateStrategy{Type: appsv1.RollingUpdateDaemonSetStrategyType}
	if err := f.engine.Client.Update(context.TODO(), ds); err != nil {
		t.Fatalf("failed to set the update strategy: %v", err)
	}

	f.editRuntime(t, func(runtime *datav1alpha1.ThinRuntime) {
		runtime.Spec.Fuse.Resources.Requests = corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("4")}
	})

	changed, err := f.engine.SyncRuntime(cruntime.ReconcileRequestContext{})
	if err != nil {
		t.Fatalf("SyncRuntime returned an error: %v", err)
	}
	if changed {
		t.Error("SyncRuntime reported a change, want it to only fix the update strategy first")
	}

	ds = f.fuseDaemonSet(t)
	if got := ds.Spec.UpdateStrategy.Type; got != appsv1.OnDeleteDaemonSetStrategyType {
		t.Errorf("update strategy = %q, want %q", got, appsv1.OnDeleteDaemonSetStrategyType)
	}
	if got := f.fuseContainer(t).Resources.Requests[corev1.ResourceCPU]; got.String() != "1" {
		t.Errorf("cpu request = %s, want the original 1 until the strategy is safe", got.String())
	}

	// The strategy is safe now, so the next reconciliation syncs the spec.
	changed, err = f.engine.SyncRuntime(cruntime.ReconcileRequestContext{})
	if err != nil {
		t.Fatalf("the second SyncRuntime returned an error: %v", err)
	}
	if !changed {
		t.Fatal("the second SyncRuntime reported no change, want the fuse template to be updated")
	}
	if got := f.fuseContainer(t).Resources.Requests[corev1.ResourceCPU]; got.String() != "4" {
		t.Errorf("cpu request = %s, want 4", got.String())
	}
}

// TestSyncRuntimeWithoutValuesConfigMap covers runtimes created with
// AnnotationDisableRuntimeHelmValueConfig, which have no last synced state to diff against.
func TestSyncRuntimeWithoutValuesConfigMap(t *testing.T) {
	f := newSyncRuntimeFixture(t, nil)

	cm := &corev1.ConfigMap{}
	if err := f.engine.Client.Get(context.TODO(), types.NamespacedName{
		Name:      f.engine.getHelmValuesConfigMapName(),
		Namespace: f.engine.namespace,
	}, cm); err != nil {
		t.Fatalf("failed to get the values configmap: %v", err)
	}
	if err := f.engine.Client.Delete(context.TODO(), cm); err != nil {
		t.Fatalf("failed to delete the values configmap: %v", err)
	}

	f.editRuntime(t, func(runtime *datav1alpha1.ThinRuntime) {
		runtime.Spec.Fuse.Resources.Requests = corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("4")}
	})

	changed, err := f.engine.SyncRuntime(cruntime.ReconcileRequestContext{})
	if err != nil {
		t.Fatalf("SyncRuntime returned an error, want it to degrade gracefully: %v", err)
	}
	if changed {
		t.Error("SyncRuntime reported a change without a values configmap")
	}
}

// TestSyncRuntimeWithoutProfile covers a dangling spec.profileName, which transformFuse cannot
// render.
func TestSyncRuntimeWithoutProfile(t *testing.T) {
	f := newSyncRuntimeFixture(t, nil)

	profile := &datav1alpha1.ThinRuntimeProfile{}
	if err := f.engine.Client.Get(context.TODO(),
		types.NamespacedName{Name: f.runtime.Spec.ThinRuntimeProfileName}, profile); err != nil {
		t.Fatalf("failed to get the profile: %v", err)
	}
	if err := f.engine.Client.Delete(context.TODO(), profile); err != nil {
		t.Fatalf("failed to delete the profile: %v", err)
	}

	changed, err := f.engine.SyncRuntime(cruntime.ReconcileRequestContext{})
	if err != nil {
		t.Fatalf("SyncRuntime returned an error, want it to degrade gracefully: %v", err)
	}
	if changed {
		t.Error("SyncRuntime reported a change without a profile")
	}
}

func envsByName(envs []corev1.EnvVar) map[string]string {
	byName := make(map[string]string, len(envs))
	for _, env := range envs {
		byName[env.Name] = env.Value
	}
	return byName
}
