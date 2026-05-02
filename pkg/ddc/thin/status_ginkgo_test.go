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

package thin

import (
	"context"
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
)

var _ = Describe("ThinEngine CheckAndUpdateRuntimeStatus", func() {
	newEngine := func(thinRuntime *datav1alpha1.ThinRuntime, objs ...k8sruntime.Object) *ThinEngine {
		allObjects := append([]k8sruntime.Object{thinRuntime}, objs...)
		return &ThinEngine{
			name:       thinRuntime.Name,
			namespace:  thinRuntime.Namespace,
			runtime:    thinRuntime,
			engineImpl: "thin",
			Client:     fake.NewFakeClientWithScheme(testScheme, allObjects...),
			Log:        fake.NullLogger(),
		}
	}

	newDataset := func(name string, mounts ...datav1alpha1.Mount) *datav1alpha1.Dataset {
		return &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: "fluid",
			},
			Spec:   datav1alpha1.DatasetSpec{Mounts: mounts},
			Status: datav1alpha1.DatasetStatus{Mounts: mounts},
		}
	}

	newRuntime := func(name string) *datav1alpha1.ThinRuntime {
		return &datav1alpha1.ThinRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:              name,
				Namespace:         "fluid",
				CreationTimestamp: metav1.NewTime(time.Now().Add(-5 * time.Minute)),
			},
			Status: datav1alpha1.RuntimeStatus{
				FusePhase: datav1alpha1.RuntimePhaseNone,
			},
		}
	}

	It("initializes fuse-only runtime status and strips mount options from dataset status", func() {
		runtime := newRuntime("fuse-only")
		engine := newEngine(
			runtime,
			newDataset("fuse-only", datav1alpha1.Mount{
				Name:           "demo",
				MountPoint:     "s3://bucket",
				Options:        map[string]string{"fs.s3a.endpoint": "endpoint"},
				EncryptOptions: []datav1alpha1.EncryptOption{{Name: "fs.s3a.secret.key"}},
			}),
		)

		ready, err := engine.CheckAndUpdateRuntimeStatus()

		Expect(err).NotTo(HaveOccurred())
		Expect(ready).To(BeTrue())

		updatedRuntime := &datav1alpha1.ThinRuntime{}
		Expect(engine.Get(context.TODO(), types.NamespacedName{Name: "fuse-only", Namespace: "fluid"}, updatedRuntime)).To(Succeed())
		Expect(updatedRuntime.Status.FusePhase).To(Equal(datav1alpha1.RuntimePhaseReady))
		Expect(updatedRuntime.Status.ValueFileConfigmap).To(Equal("fuse-only-thin-values"))
		Expect(updatedRuntime.Status.SetupDuration).NotTo(BeEmpty())
		Expect(updatedRuntime.Status.CacheStates).To(Equal(common.CacheStateList{
			common.CacheCapacity:        "N/A",
			common.CachedPercentage:     "N/A",
			common.Cached:               "N/A",
			common.CacheHitRatio:        "N/A",
			common.CacheThroughputRatio: "N/A",
		}))
		Expect(updatedRuntime.Status.Mounts).To(HaveLen(1))
		Expect(updatedRuntime.Status.Mounts[0].MountPoint).To(Equal("s3://bucket"))
		Expect(updatedRuntime.Status.Mounts[0].Options).To(BeNil())
		Expect(updatedRuntime.Status.Mounts[0].EncryptOptions).To(BeNil())
		_, condition := utils.GetRuntimeCondition(updatedRuntime.Status.Conditions, datav1alpha1.RuntimeFusesInitialized)
		Expect(condition).NotTo(BeNil())
	})

	It("updates cache affinity and waits for ready workers before marking the runtime ready", func() {
		runtime := newRuntime("workers")
		runtime.Spec.Replicas = 2
		runtime.Spec.Worker.Enabled = true

		workers := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "workers-worker",
				Namespace: "fluid",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: ptr.To[int32](2),
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						NodeSelector: map[string]string{"topology.kubernetes.io/zone": "zone-a"},
					},
				},
			},
			Status: appsv1.StatefulSetStatus{ReadyReplicas: 0},
		}

		engine := newEngine(runtime, newDataset("workers", datav1alpha1.Mount{Name: "demo", MountPoint: "oss://bucket"}), workers)

		ready, err := engine.CheckAndUpdateRuntimeStatus()

		Expect(err).NotTo(HaveOccurred())
		Expect(ready).To(BeTrue())

		updatedRuntime := &datav1alpha1.ThinRuntime{}
		Expect(engine.Get(context.TODO(), types.NamespacedName{Name: "workers", Namespace: "fluid"}, updatedRuntime)).To(Succeed())
		Expect(updatedRuntime.Status.FusePhase).To(Equal(datav1alpha1.RuntimePhaseReady))
		Expect(updatedRuntime.Status.CacheAffinity).NotTo(BeNil())
		Expect(updatedRuntime.Status.CacheAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms).To(HaveLen(1))
		Expect(updatedRuntime.Status.CacheAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions).To(ContainElement(corev1.NodeSelectorRequirement{
			Key:      "topology.kubernetes.io/zone",
			Operator: corev1.NodeSelectorOpIn,
			Values:   []string{"zone-a"},
		}))
	})
})
