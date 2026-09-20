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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	workloadv1alpha1 "github.com/fluid-cloudnative/advanced-statefulset/api/workload/v1alpha1"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// This file is the executable form of the pod-template composition table tracked in #6185.
//
// A component's pod template is composed from three layers:
//
//	L1  CacheRuntimeClass.topology.<component>.template   (a full PodTemplateSpec)
//	L2  CacheRuntime.spec.*                               (runtime level, all components)
//	L3  CacheRuntime.spec.<component>.*                   (component level)
//
// Every row below states what each layer declares and what the component must end up
// with. Each row is then replayed through BOTH code paths that compose those layers:
//
//	create  the transform path, which builds the workload from all three layers at once
//	update  the sync path, which patches a workload created from L1 alone
//
// The two paths are written independently (transform_common.go and
// advanced_statefulset_manager.go) and are expected to agree. A row that passes on one
// path and fails on the other is a divergence bug, which is the class #6185 is about.
//
// Rows carrying a knownBug tag pin CURRENT behaviour, not intended behaviour. Do not
// read them as an endorsement: each one names the issue that will change it, and the
// comment states what the row should say once that issue is fixed.

// --- the table ------------------------------------------------------------------

type layerCase struct {
	// field names the row in the table. Several rows may share a field.
	field string
	desc  string

	l1 func(*corev1.PodTemplateSpec)
	l2 func(*datav1alpha1.CacheRuntimeSpec)
	l3 func(*datav1alpha1.CacheRuntimeWorkerSpec)

	// want asserts on the pod template the component ends up with, whichever path produced
	// it. The whole template is handed over, not just the PodSpec, because podMetadata and
	// imagePullSecrets live on different halves of it.
	want func(g Gomega, tmpl corev1.PodTemplateSpec)

	// onlyPaths restricts the row to the named paths. Empty means both. Use it only for
	// fields a path genuinely cannot carry, and say why -- not to silence a divergence.
	onlyPaths []string

	// knownBug names the issue this row's expectation is wrong under. Non-empty means the
	// row pins current behaviour so a change to it is visible in review.
	knownBug string
}

func compositionTable() []layerCase {
	return []layerCase{
		// -- podMetadata: union, later layer wins a key ------------------------------
		{
			field: "podMetadata.labels",
			desc:  "L2 alone reaches the pod",
			l2: func(s *datav1alpha1.CacheRuntimeSpec) {
				s.PodMetadata.Labels = map[string]string{"from": "runtime"}
			},
			want: func(g Gomega, t corev1.PodTemplateSpec) {
				g.Expect(t.Labels).To(HaveKeyWithValue("from", "runtime"))
			},
			onlyPaths: []string{pathCreate}, // sync path does not patch pod metadata
		},

		// -- imagePullSecrets: merge by name -----------------------------------------
		{
			field: "imagePullSecrets",
			desc:  "L1 and L2 merge, a name declared twice appears once",
			l1: func(t *corev1.PodTemplateSpec) {
				t.Spec.ImagePullSecrets = []corev1.LocalObjectReference{{Name: "tmpl"}, {Name: "shared"}}
			},
			l2: func(s *datav1alpha1.CacheRuntimeSpec) {
				s.ImagePullSecrets = []corev1.LocalObjectReference{{Name: "shared"}, {Name: "rt"}}
			},
			want: func(g Gomega, t corev1.PodTemplateSpec) {
				g.Expect(t.Spec.ImagePullSecrets).To(ConsistOf(
					corev1.LocalObjectReference{Name: "tmpl"},
					corev1.LocalObjectReference{Name: "shared"},
					corev1.LocalObjectReference{Name: "rt"}))
			},
			onlyPaths: []string{pathCreate},
		},

		// -- resources: the row #6173 and #6185 turn on ------------------------------
		{
			field: "resources",
			desc:  "L3 restates only limits.memory over an L1 that declares four keys",
			l1: func(t *corev1.PodTemplateSpec) {
				t.Spec.Containers[0].Resources = corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("1"),
						corev1.ResourceMemory: resource.MustParse("2Gi"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("2"),
						corev1.ResourceMemory: resource.MustParse("4Gi"),
					},
				}
			},
			l3: func(w *datav1alpha1.CacheRuntimeWorkerSpec) {
				w.Resources = corev1.ResourceRequirements{
					Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("8Gi")},
				}
			},
			// CURRENT: the whole struct is replaced, so the three keys L3 did not restate
			// are lost and the container ends up with no CPU request or limit at all.
			// AFTER #6173: requests {cpu 1, memory 2Gi}, limits {cpu 2, memory 8Gi}.
			knownBug: "#6173",
			want: func(g Gomega, t corev1.PodTemplateSpec) {
				r := t.Spec.Containers[0].Resources
				g.Expect(r.Limits).To(HaveKeyWithValue(corev1.ResourceMemory, resource.MustParse("8Gi")))
				g.Expect(r.Limits).NotTo(HaveKey(corev1.ResourceCPU))
				g.Expect(r.Requests).To(BeEmpty())
			},
		},

		// -- runtimeVersion: the guard #6178 is about --------------------------------
		{
			field: "runtimeVersion",
			desc:  "L3 sets imageTag only, over an L1 that names an image",
			l1: func(t *corev1.PodTemplateSpec) {
				t.Spec.Containers[0].Image = "fluid/cache:v1"
			},
			l3: func(w *datav1alpha1.CacheRuntimeWorkerSpec) {
				w.RuntimeVersion = datav1alpha1.VersionSpec{ImageTag: "v2"}
			},
			// CURRENT: the guard wants both image and imageTag, so the tag is dropped.
			// AFTER #6178: fluid/cache:v2.
			knownBug: "#6178",
			want: func(g Gomega, t corev1.PodTemplateSpec) {
				g.Expect(t.Spec.Containers[0].Image).To(Equal("fluid/cache:v1"))
			},
		},

		// -- args vs env: replace on one, append on the other ------------------------
		{
			field: "args",
			desc:  "L3 replaces the template args wholesale",
			l1: func(t *corev1.PodTemplateSpec) {
				t.Spec.Containers[0].Args = []string{"--from-template"}
			},
			l3: func(w *datav1alpha1.CacheRuntimeWorkerSpec) {
				w.Args = []string{"--from-component"}
			},
			// Intended: the CRD marks args +listType=atomic, so replacing is correct.
			want: func(g Gomega, t corev1.PodTemplateSpec) {
				g.Expect(t.Spec.Containers[0].Args).To(Equal([]string{"--from-component"}))
			},
			onlyPaths: []string{pathCreate}, // args are not an in-place update field
		},
		{
			field: "env",
			desc:  "L3 restates a name the template already sets",
			l1: func(t *corev1.PodTemplateSpec) {
				t.Spec.Containers[0].Env = []corev1.EnvVar{{Name: "SHARED", Value: "template"}}
			},
			l3: func(w *datav1alpha1.CacheRuntimeWorkerSpec) {
				w.Env = []corev1.EnvVar{{Name: "SHARED", Value: "component"}}
			},
			// CURRENT: appended without dedup, so SHARED appears twice. The CRD marks env
			// +patchStrategy=merge +patchMergeKey=name, which says it should appear once.
			// Kubernetes takes the last value, so the effective value is already correct.
			knownBug: "#6185 (env/annotation mismatch)",
			want: func(g Gomega, t corev1.PodTemplateSpec) {
				var shared []string
				for _, e := range t.Spec.Containers[0].Env {
					if e.Name == "SHARED" {
						shared = append(shared, e.Value)
					}
				}
				g.Expect(shared).To(Equal([]string{"template", "component"}))
			},
			onlyPaths: []string{pathCreate},
		},

		// -- nodeSelector: implementation and annotation disagree --------------------
		{
			field: "nodeSelector",
			desc:  "L3 sets a key the template does not",
			l1: func(t *corev1.PodTemplateSpec) {
				t.Spec.NodeSelector = map[string]string{"tier": "template"}
			},
			l3: func(w *datav1alpha1.CacheRuntimeWorkerSpec) {
				w.NodeSelector = map[string]string{"zone": "a"}
			},
			// CURRENT: union, so the template key survives. The CRD marks nodeSelector
			// +mapType=atomic, which says the whole map should be replaced.
			knownBug: "#6185 (nodeSelector/annotation mismatch)",
			want: func(g Gomega, t corev1.PodTemplateSpec) {
				g.Expect(t.Spec.NodeSelector).To(HaveKeyWithValue("tier", "template"))
				g.Expect(t.Spec.NodeSelector).To(HaveKeyWithValue("zone", "a"))
			},
			onlyPaths: []string{pathCreate},
		},
	}
}

// --- the two paths --------------------------------------------------------------

const (
	pathCreate = "create (transform)"
	pathUpdate = "update (sync)"
)

type compositionPath struct {
	name string
	// run applies the three layers and returns the pod template the component ends up with.
	run func(g Gomega, tc layerCase) corev1.PodTemplateSpec
}

func baseTemplate() corev1.PodTemplateSpec {
	return corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{Name: "worker", Image: "fluid/cache:v1"}},
		},
	}
}

func layersOf(tc layerCase) (corev1.PodTemplateSpec, datav1alpha1.CacheRuntimeSpec) {
	tmpl := baseTemplate()
	if tc.l1 != nil {
		tc.l1(&tmpl)
	}
	runtimeSpec := datav1alpha1.CacheRuntimeSpec{RuntimeClassName: "test-class"}
	if tc.l2 != nil {
		tc.l2(&runtimeSpec)
	}
	worker := datav1alpha1.CacheRuntimeWorkerSpec{Replicas: 1}
	if tc.l3 != nil {
		tc.l3(&worker)
	}
	runtimeSpec.Worker = worker
	return tmpl, runtimeSpec
}

// createPath composes all three layers the way workload creation does.
var createPath = compositionPath{
	name: pathCreate,
	run: func(g Gomega, tc layerCase) corev1.PodTemplateSpec {
		tmpl, runtimeSpec := layersOf(tc)
		e := &CacheEngine{name: "test", namespace: "default"}

		value, err := e.initComponentValue(common.ComponentTypeWorker,
			&datav1alpha1.RuntimeComponentDefinition{Template: tmpl}, nil, runtimeSpec.Worker.Replicas)
		g.Expect(err).NotTo(HaveOccurred())

		e.transformComponentPodTemplate(runtimeSpec, runtimeSpec.Worker.RuntimeComponentCommonSpec,
			&datav1alpha1.Dataset{ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"}}, value)

		return value.PodTemplateSpec
	},
}

// updatePath seeds a workload built from L1 alone -- the state a CacheRuntime that set
// nothing would have produced -- then applies L2 and L3 through the sync path and reads
// the workload back. The end state must match the create path's.
var updatePath = compositionPath{
	name: pathUpdate,
	run: func(g Gomega, tc layerCase) corev1.PodTemplateSpec {
		tmpl, runtimeSpec := layersOf(tc)
		name := common.GetCacheComponentName("test", common.ComponentTypeWorker)

		seeded := baseTemplate()
		if tc.l1 != nil {
			tc.l1(&seeded)
		}
		replicas := int32(1)
		asts := &workloadv1alpha1.AdvancedStatefulSet{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "default"},
			Spec: workloadv1alpha1.AdvancedStatefulSetSpec{
				Replicas: &replicas,
				Template: seeded,
			},
		}

		runtimeObj := &datav1alpha1.CacheRuntime{
			ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
			Spec:       runtimeSpec,
		}
		runtimeClass := &datav1alpha1.CacheRuntimeClass{
			ObjectMeta: metav1.ObjectMeta{Name: "test-class"},
			Topology: &datav1alpha1.RuntimeTopology{
				Worker: &datav1alpha1.RuntimeComponentDefinition{Template: tmpl},
			},
		}

		client := fake.NewClientBuilder().
			WithScheme(CacheEngineTestScheme).
			WithObjects(asts, runtimeObj, runtimeClass).
			Build()

		e := &CacheEngine{
			name:      "test",
			namespace: "default",
			Client:    client,
			Log:       ctrl.Log.WithName("composition-matrix"),
		}

		g.Expect(e.syncRuntimeSpec(cruntime.ReconcileRequestContext{}, runtimeObj, runtimeClass)).To(Succeed())

		got := &workloadv1alpha1.AdvancedStatefulSet{}
		g.Expect(client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: "default"}, got)).To(Succeed())
		return got.Spec.Template
	},
}

// --- the runner -----------------------------------------------------------------

var _ = Describe("CacheRuntime pod template composition matrix",
	Label("pkg.ddc.cache.engine.composition_matrix_test.go"), func() {

		for _, path := range []compositionPath{createPath, updatePath} {
			path := path

			Describe(path.name, func() {
				for _, tc := range compositionTable() {
					tc := tc

					if !runsOn(tc, path.name) {
						continue
					}

					name := tc.field + ": " + tc.desc
					if tc.knownBug != "" {
						name += " [pins current behaviour, " + tc.knownBug + "]"
					}

					It(name, func() {
						tc.want(Default, path.run(Default, tc))
					})
				}
			})
		}
	})

func runsOn(tc layerCase, path string) bool {
	if len(tc.onlyPaths) == 0 {
		return true
	}
	for _, p := range tc.onlyPaths {
		if p == path {
			return true
		}
	}
	return false
}
