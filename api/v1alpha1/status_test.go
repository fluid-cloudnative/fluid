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

package v1alpha1

import (
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("status deep copy helpers", func() {
	Describe("RuntimeStatus.DeepCopy", func() {
		It("preserves fields while cloning nested references", func() {
			mountTime := metav1.NewTime(time.Unix(1700000000, 0))
			status := &RuntimeStatus{
				ValueFileConfigmap: "values",
				WorkerPhase:        RuntimePhaseReady,
				Selector:           "app=runtime",
				Conditions: []RuntimeCondition{{
					Type:               RuntimeWorkersReady,
					Status:             corev1.ConditionTrue,
					Reason:             RuntimeWorkersReadyReason,
					LastTransitionTime: mountTime,
				}},
				MountTime: &mountTime,
				Mounts: []Mount{{
					MountPoint: "s3://bucket/data",
				}},
				CacheAffinity: &corev1.NodeAffinity{},
			}

			copy := status.DeepCopy()

			Expect(copy).NotTo(BeNil())
			Expect(copy).NotTo(BeIdenticalTo(status))
			Expect(copy.WorkerPhase).To(Equal(RuntimePhaseReady))
			Expect(copy.Conditions).To(HaveLen(1))
			Expect(copy.Conditions[0].Reason).To(Equal(RuntimeWorkersReadyReason))
			Expect(copy.MountTime).NotTo(BeNil())
			Expect(copy.MountTime).NotTo(BeIdenticalTo(status.MountTime))
			Expect(copy.MountTime.Time).To(Equal(status.MountTime.Time))
			Expect(copy.Mounts).To(HaveLen(1))
			Expect(copy.Mounts[0].MountPoint).To(Equal("s3://bucket/data"))

			copy.Mounts[0].MountPoint = "changed"
			Expect(status.Mounts[0].MountPoint).To(Equal("s3://bucket/data"))
		})
	})

	Describe("CacheRuntimeStatus.DeepCopy", func() {
		It("deep copies runtime component and mount point status", func() {
			mountTime := metav1.NewTime(time.Unix(1700000100, 0))
			status := &CacheRuntimeStatus{
				Selector: "app=cache",
				RuntimeComponentStatusCollection: RuntimeComponentStatusCollection{
					Worker: RuntimeComponentStatus{
						Phase:           RuntimePhasePartialReady,
						ReadyReplicas:   1,
						DesiredReplicas: 2,
					},
				},
				MountPoints: []MountPointStatus{{
					Mount:     Mount{MountPoint: "/data/cache"},
					MountTime: &mountTime,
				}},
			}

			copy := status.DeepCopy()

			Expect(copy).NotTo(BeNil())
			Expect(copy).NotTo(BeIdenticalTo(status))
			Expect(copy.Worker.Phase).To(Equal(RuntimePhasePartialReady))
			Expect(copy.MountPoints).To(HaveLen(1))
			Expect(copy.MountPoints[0].Mount.MountPoint).To(Equal("/data/cache"))
			Expect(copy.MountPoints[0].MountTime).NotTo(BeIdenticalTo(status.MountPoints[0].MountTime))

			copy.MountPoints[0].Mount.MountPoint = "/mutated"
			Expect(status.MountPoints[0].Mount.MountPoint).To(Equal("/data/cache"))
		})
	})

	Describe("OperationStatus.DeepCopy", func() {
		It("preserves operation metadata and nested pointers", func() {
			complete := true
			lastScheduleTime := metav1.NewTime(time.Unix(1700000200, 0))
			status := &OperationStatus{
				Phase:    common.PhaseComplete,
				Duration: "10s",
				Conditions: []Condition{{
					Status:             corev1.ConditionTrue,
					Reason:             "Completed",
					LastTransitionTime: lastScheduleTime,
				}},
				Infos:            map[string]string{"path": "/tmp/data"},
				LastScheduleTime: &lastScheduleTime,
				WaitingFor: WaitingStatus{
					OperationComplete: &complete,
				},
				NodeAffinity: &corev1.NodeAffinity{},
			}

			copy := status.DeepCopy()

			Expect(copy).NotTo(BeNil())
			Expect(copy).NotTo(BeIdenticalTo(status))
			Expect(copy.Phase).To(Equal(common.PhaseComplete))
			Expect(copy.Duration).To(Equal("10s"))
			Expect(copy.Infos).To(Equal(map[string]string{"path": "/tmp/data"}))
			Expect(copy.LastScheduleTime).NotTo(BeIdenticalTo(status.LastScheduleTime))
			Expect(copy.WaitingFor.OperationComplete).NotTo(BeNil())
			Expect(*copy.WaitingFor.OperationComplete).To(BeTrue())

			copy.Infos["path"] = "/tmp/other"
			Expect(status.Infos["path"]).To(Equal("/tmp/data"))
		})
	})
})
