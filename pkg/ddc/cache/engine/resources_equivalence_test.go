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
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// The creation path (transformComponentPodTemplate) and the sync path
// (desiredComponentResources) both resolve a component's resources through
// mergeResourceRequirements. They must agree for every combination of template and
// CacheRuntime resources: if they did not, the first reconcile after creation would patch
// the workload it had just rendered, and every reconcile after that would keep patching.
//
// Each row below is checked twice:
//
//	resolve   the two resolvers return the same value, compared the way updateResources
//	          compares it (equality.Semantic); the sync side charges the tiered store
//	          memory quota from the rendered workload, as syncRuntimeSpec does
//	converge  a workload rendered by the creation path is left untouched by syncRuntimeSpec

type resourcesEquivalenceCase struct {
	desc        string
	template    corev1.ResourceRequirements
	runtime     corev1.ResourceRequirements
	tieredStore datav1alpha1.RuntimeTieredStore
	// For a tiered store row: the quota the creation path must charge and the memory
	// it must end up with, so the row cannot pass with the quota silently dropped.
	wantQuota, wantMemoryRequest, wantMemoryLimit string
}

func resourcesEquivalenceTable() []resourcesEquivalenceCase {
	full := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("1"),
			corev1.ResourceMemory: resource.MustParse("2Gi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("2"),
			corev1.ResourceMemory: resource.MustParse("4Gi"),
		},
	}

	return []resourcesEquivalenceCase{
		{
			desc: "neither side declares resources",
		},
		{
			desc:     "only the template declares resources",
			template: full,
		},
		{
			desc: "only the CacheRuntime declares resources",
			runtime: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("8Gi")},
			},
		},
		{
			// #6173
			desc:     "the CacheRuntime names one key of a full template",
			template: full,
			runtime: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("8Gi")},
			},
		},
		{
			desc: "the CacheRuntime names a key the template does not declare",
			template: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("4Gi")},
			},
			runtime: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("500m")},
			},
		},
		{
			// The nil/empty boundary on the overlay side: an empty list names no key.
			desc: "the CacheRuntime declares an empty list",
			template: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("4Gi")},
			},
			runtime: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{},
			},
		},
		{
			// The same boundary on the baseline side.
			desc: "the template declares an empty list",
			template: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{},
			},
			runtime: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("8Gi")},
			},
		},
		{
			desc: "both sides declare claims",
			template: corev1.ResourceRequirements{
				Claims: []corev1.ResourceClaim{{Name: "gpu"}, {Name: "nic"}},
			},
			runtime: corev1.ResourceRequirements{
				Claims: []corev1.ResourceClaim{{Name: "nic"}, {Name: "fpga"}},
			},
		},
		{
			// The creation path adds the quota to the merged memory; the sync path must
			// add the same quota instead of reverting to the bare merge.
			desc:     "a Memory-medium emptyDir tier charges the merged memory",
			template: full,
			runtime: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("8Gi")},
			},
			tieredStore: datav1alpha1.RuntimeTieredStore{
				Levels: []datav1alpha1.RuntimeTieredStoreLevel{{
					EmptyDir: &datav1alpha1.EmptyDirMediumSource{
						Quota:  resource.MustParse("2Gi"),
						Medium: corev1.StorageMediumMemory,
					},
				}},
			},
			wantQuota:         "2Gi",
			wantMemoryRequest: "4Gi",
			wantMemoryLimit:   "10Gi",
		},
		{
			desc:     "a process memory tier charges the merged memory",
			template: full,
			runtime: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("3Gi")},
			},
			tieredStore: datav1alpha1.RuntimeTieredStore{
				Levels: []datav1alpha1.RuntimeTieredStoreLevel{{
					ProcessMemory: &datav1alpha1.ProcessMemoryMediumSource{Quota: resource.MustParse("1Gi")},
				}},
			},
			wantQuota:         "1Gi",
			wantMemoryRequest: "4Gi",
			wantMemoryLimit:   "5Gi",
		},
	}
}

func workerTemplateWith(resources corev1.ResourceRequirements) corev1.PodTemplateSpec {
	return corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name:      "worker",
				Image:     "fluid/cache:v1",
				Resources: *resources.DeepCopy(),
			}},
		},
	}
}

// renderWorker runs the creation path and returns the worker pod template it renders.
func renderWorker(g Gomega, tc resourcesEquivalenceCase) corev1.PodTemplateSpec {
	e := &CacheEngine{name: "test", namespace: "default"}
	runtimeSpec := datav1alpha1.CacheRuntimeSpec{
		RuntimeClassName: "test-class",
		Worker: datav1alpha1.CacheRuntimeWorkerSpec{
			Replicas:                   1,
			RuntimeComponentCommonSpec: datav1alpha1.RuntimeComponentCommonSpec{Resources: *tc.runtime.DeepCopy()},
		},
	}

	value, err := e.initComponentValue(common.ComponentTypeWorker,
		&datav1alpha1.RuntimeComponentDefinition{Template: workerTemplateWith(tc.template)}, nil, 1)
	g.Expect(err).NotTo(HaveOccurred())

	e.transformComponentPodTemplate(runtimeSpec, runtimeSpec.Worker.RuntimeComponentCommonSpec,
		&datav1alpha1.Dataset{ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"}}, value)
	g.Expect(e.TransformRuntimeTieredStore(tc.tieredStore.DeepCopy(), &value.PodTemplateSpec.Spec)).To(Succeed())

	return value.PodTemplateSpec
}

var _ = Describe("CacheRuntime resources creation/sync equivalence",
	Label("pkg.ddc.cache.engine.resources_equivalence_test.go"), func() {

		for _, tc := range resourcesEquivalenceTable() {
			tc := tc

			Describe(tc.desc, func() {
				if tc.wantQuota != "" {
					It("charges the tiered store quota on the creation path", func() {
						pod := renderWorker(Default, tc)
						quota := chargedTieredStoreMemoryQuota(&pod.Spec)
						resources := pod.Spec.Containers[0].Resources

						Expect(quota.Cmp(resource.MustParse(tc.wantQuota))).To(BeZero(), "quota %s", quota.String())
						Expect(resources.Requests.Memory().Cmp(resource.MustParse(tc.wantMemoryRequest))).To(BeZero(),
							"memory request %s", resources.Requests.Memory().String())
						Expect(resources.Limits.Memory().Cmp(resource.MustParse(tc.wantMemoryLimit))).To(BeZero(),
							"memory limit %s", resources.Limits.Memory().String())
					})
				}

				It("resolves to the same resources on both paths", func() {
					renderedPod := renderWorker(Default, tc)
					rendered := renderedPod.Spec.Containers[0].Resources
					merged := desiredComponentResources(tc.runtime,
						&datav1alpha1.RuntimeComponentDefinition{Template: workerTemplateWith(tc.template)})
					desired := withTieredStoreMemoryQuota(*merged, chargedTieredStoreMemoryQuota(&renderedPod.Spec))

					Expect(equality.Semantic.DeepEqual(rendered, desired)).To(BeTrue(),
						"creation rendered %+v, sync wants %+v", rendered, desired)
				})

				It("leaves a freshly created workload untouched on sync", func() {
					name := common.GetCacheComponentName("test", common.ComponentTypeWorker)
					replicas := int32(1)
					asts := &workloadv1alpha1.AdvancedStatefulSet{
						ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "default"},
						Spec: workloadv1alpha1.AdvancedStatefulSetSpec{
							Replicas: &replicas,
							Template: renderWorker(Default, tc),
						},
					}
					runtimeObj := &datav1alpha1.CacheRuntime{
						ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
						Spec: datav1alpha1.CacheRuntimeSpec{
							RuntimeClassName: "test-class",
							Worker: datav1alpha1.CacheRuntimeWorkerSpec{
								Replicas:                   1,
								RuntimeComponentCommonSpec: datav1alpha1.RuntimeComponentCommonSpec{Resources: *tc.runtime.DeepCopy()},
								TieredStore:                *tc.tieredStore.DeepCopy(),
							},
						},
					}
					runtimeClass := &datav1alpha1.CacheRuntimeClass{
						ObjectMeta: metav1.ObjectMeta{Name: "test-class"},
						Topology: &datav1alpha1.RuntimeTopology{
							Worker: &datav1alpha1.RuntimeComponentDefinition{Template: workerTemplateWith(tc.template)},
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
						Log:       ctrl.Log.WithName("resources-equivalence"),
					}

					key := types.NamespacedName{Name: name, Namespace: "default"}
					created := &workloadv1alpha1.AdvancedStatefulSet{}
					Expect(client.Get(context.TODO(), key, created)).To(Succeed())

					Expect(e.syncRuntimeSpec(cruntime.ReconcileRequestContext{}, runtimeObj, runtimeClass)).To(Succeed())

					synced := &workloadv1alpha1.AdvancedStatefulSet{}
					Expect(client.Get(context.TODO(), key, synced)).To(Succeed())
					Expect(synced.ResourceVersion).To(Equal(created.ResourceVersion),
						"sync patched a workload the creation path had just rendered")
				})
			})
		}
	})
