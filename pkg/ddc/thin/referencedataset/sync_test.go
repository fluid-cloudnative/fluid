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

package referencedataset

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("ReferenceDatasetEngine Sync Tests", Label("pkg.ddc.thin.referencedataset.sync_test.go"), func() {
	Describe("Sync", func() {
		var (
			fakeClient    client.Client
			refRuntimeObj datav1alpha1.ThinRuntime
			refDatasetObj datav1alpha1.Dataset
			datasetObj    datav1alpha1.Dataset
			alluxioRt     datav1alpha1.AlluxioRuntime
			testScheme    *runtime.Scheme
		)

		BeforeEach(func() {
			testScheme = runtime.NewScheme()
			_ = v1.AddToScheme(testScheme)
			_ = datav1alpha1.AddToScheme(testScheme)
			_ = appsv1.AddToScheme(testScheme)

			datasetObj = datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "done",
					Namespace: "big-data",
				},
				Status: datav1alpha1.DatasetStatus{
					Runtimes: []datav1alpha1.Runtime{
						{
							Name:      "done",
							Namespace: "big-data",
							Type:      common.AlluxioRuntime,
						},
					},
					DatasetRef: []string{
						"fluid/hbase",
						"fluid/test",
					},
					UfsTotal: "100Gi",
				},
			}

			alluxioRt = datav1alpha1.AlluxioRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "done",
					Namespace: "big-data",
				},
			}

			refRuntimeObj = datav1alpha1.ThinRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "fluid",
				},
			}

			refDatasetObj = datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{
							MountPoint: "dataset://big-data/done",
						},
					},
				},
			}

			testObjs := []runtime.Object{&datasetObj, &refDatasetObj, &alluxioRt, &refRuntimeObj}
			fakeClient = fake.NewFakeClientWithScheme(testScheme, testObjs...)
		})

		Context("when syncing reference dataset", func() {
			It("should sync successfully and update dataset status", func() {
				e := &ReferenceDatasetEngine{
					Id:                "reference-engine",
					Client:            fakeClient,
					Log:               fake.NullLogger(),
					name:              refRuntimeObj.GetName(),
					namespace:         refRuntimeObj.GetNamespace(),
					timeOfLastSync:    time.Now().Add(-defaultSyncRetryDuration),
					syncRetryDuration: defaultSyncRetryDuration,
				}

				ctx := cruntime.ReconcileRequestContext{}
				err := e.Sync(ctx)
				Expect(err).NotTo(HaveOccurred())

				updatedRefDataset := &datav1alpha1.Dataset{}
				err = fakeClient.Get(context.TODO(), types.NamespacedName{
					Namespace: refDatasetObj.Namespace, Name: refDatasetObj.Name,
				}, updatedRefDataset)
				Expect(err).NotTo(HaveOccurred())

				// Check updated status
				Expect(updatedRefDataset.Status.UfsTotal).To(Equal(datasetObj.Status.UfsTotal))
				Expect(len(updatedRefDataset.Status.DatasetRef)).To(Equal(0))

				boundRuntimes := updatedRefDataset.Status.Runtimes
				Expect(len(boundRuntimes)).To(Equal(1))

				boundRuntime := boundRuntimes[0]
				Expect(boundRuntime.Type).To(Equal(common.ThinRuntime))
				Expect(boundRuntime.Name).To(Equal(refRuntimeObj.Name))
				Expect(boundRuntime.Namespace).To(Equal(refRuntimeObj.Namespace))
			})
		})
	})

	Describe("Sync with CacheRuntime", func() {
		var (
			fakeClient      client.Client
			refRuntimeObj   datav1alpha1.ThinRuntime
			refDatasetObj   datav1alpha1.Dataset
			datasetObj      datav1alpha1.Dataset
			cacheRuntimeObj datav1alpha1.CacheRuntime
			testScheme      *runtime.Scheme
		)

		BeforeEach(func() {
			testScheme = runtime.NewScheme()
			_ = v1.AddToScheme(testScheme)
			_ = datav1alpha1.AddToScheme(testScheme)
			_ = appsv1.AddToScheme(testScheme)

			// Create physical dataset with CacheRuntime
			datasetObj = datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cache-runtime",
					Namespace: "cache-ns",
				},
				Status: datav1alpha1.DatasetStatus{
					Runtimes: []datav1alpha1.Runtime{
						{
							Name:      "cache-runtime",
							Namespace: "cache-ns",
							Type:      common.CacheRuntime,
						},
					},
					UfsTotal: "500Gi",
					Phase:    datav1alpha1.BoundDatasetPhase,
				},
			}

			// Create CacheRuntime with status
			cacheRuntimeObj = datav1alpha1.CacheRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cache-runtime",
					Namespace: "cache-ns",
				},
				Spec: datav1alpha1.CacheRuntimeSpec{
					RuntimeClassName: "curvine",
				},
				Status: datav1alpha1.CacheRuntimeStatus{
					ValueFile:     "/tmp/cache-values.yaml",
					SetupDuration: "1m30s",
					Selector:      "cache=curvine",
					RuntimeComponentStatusCollection: datav1alpha1.RuntimeComponentStatusCollection{
						Master: datav1alpha1.RuntimeComponentStatus{
							Phase:             datav1alpha1.RuntimePhaseReady,
							Reason:            "MasterReady",
							DesiredReplicas:   1,
							CurrentReplicas:   1,
							ReadyReplicas:     1,
							AvailableReplicas: 1,
						},
						Worker: datav1alpha1.RuntimeComponentStatus{
							Phase:               datav1alpha1.RuntimePhaseReady,
							Reason:              "WorkersReady",
							DesiredReplicas:     2,
							CurrentReplicas:     2,
							ReadyReplicas:       2,
							AvailableReplicas:   2,
							UnavailableReplicas: 0,
						},
						Client: datav1alpha1.RuntimeComponentStatus{
							Phase:               datav1alpha1.RuntimePhaseReady,
							Reason:              "FusesReady",
							DesiredReplicas:     2,
							CurrentReplicas:     2,
							ReadyReplicas:       2,
							AvailableReplicas:   2,
							UnavailableReplicas: 0,
						},
					},
					Conditions: []datav1alpha1.RuntimeCondition{
						{
							Type:   datav1alpha1.RuntimeMasterReady,
							Status: v1.ConditionTrue,
						},
						{
							Type:   datav1alpha1.RuntimeWorkersReady,
							Status: v1.ConditionTrue,
						},
					},
				},
			}

			// Create ThinRuntime (reference dataset runtime)
			refRuntimeObj = datav1alpha1.ThinRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ref-dataset",
					Namespace: "ref-ns",
				},
			}

			// Create reference dataset that points to the physical dataset
			refDatasetObj = datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ref-dataset",
					Namespace: "ref-ns",
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{
							MountPoint: "dataset://cache-ns/cache-runtime",
							Name:       "cached-data",
						},
					},
				},
			}

			testObjs := []runtime.Object{&datasetObj, &cacheRuntimeObj, &refDatasetObj, &refRuntimeObj}
			fakeClient = fake.NewFakeClientWithScheme(testScheme, testObjs...)
		})

		Context("when syncing reference dataset with CacheRuntime backend", func() {
			It("should sync successfully and update ThinRuntime status from CacheRuntime", func() {
				e := &ReferenceDatasetEngine{
					Id:                "cache-ref-engine",
					Client:            fakeClient,
					Log:               fake.NullLogger(),
					name:              refDatasetObj.GetName(),
					namespace:         refDatasetObj.GetNamespace(),
					timeOfLastSync:    time.Now().Add(-defaultSyncRetryDuration),
					syncRetryDuration: defaultSyncRetryDuration,
				}

				ctx := cruntime.ReconcileRequestContext{}
				err := e.Sync(ctx)
				Expect(err).NotTo(HaveOccurred())

				// Verify reference dataset status was updated
				updatedRefDataset := &datav1alpha1.Dataset{}
				err = fakeClient.Get(context.TODO(), types.NamespacedName{
					Namespace: refDatasetObj.Namespace,
					Name:      refDatasetObj.Name,
				}, updatedRefDataset)
				Expect(err).NotTo(HaveOccurred())

				// Check dataset status synced from physical dataset
				Expect(updatedRefDataset.Status.UfsTotal).To(Equal(datasetObj.Status.UfsTotal))
				Expect(updatedRefDataset.Status.Phase).To(Equal(datav1alpha1.BoundDatasetPhase))

				// Verify ThinRuntime status was updated from CacheRuntime
				updatedThinRuntime := &datav1alpha1.ThinRuntime{}
				err = fakeClient.Get(context.TODO(), types.NamespacedName{
					Namespace: refRuntimeObj.Namespace,
					Name:      refRuntimeObj.Name,
				}, updatedThinRuntime)
				Expect(err).NotTo(HaveOccurred())

				// Verify status conversion from CacheRuntime to RuntimeStatus
				Expect(updatedThinRuntime.Status.ValueFileConfigmap).To(Equal("/tmp/cache-values.yaml"))
				Expect(updatedThinRuntime.Status.MasterPhase).To(Equal(datav1alpha1.RuntimePhaseReady))
				Expect(updatedThinRuntime.Status.WorkerPhase).To(Equal(datav1alpha1.RuntimePhaseReady))
				Expect(updatedThinRuntime.Status.FusePhase).To(Equal(datav1alpha1.RuntimePhaseReady))

				// Verify replica counts
				Expect(updatedThinRuntime.Status.DesiredMasterNumberScheduled).To(Equal(int32(1)))
				Expect(updatedThinRuntime.Status.CurrentMasterNumberScheduled).To(Equal(int32(1)))
				Expect(updatedThinRuntime.Status.MasterNumberReady).To(Equal(int32(1)))

				Expect(updatedThinRuntime.Status.DesiredWorkerNumberScheduled).To(Equal(int32(2)))
				Expect(updatedThinRuntime.Status.CurrentWorkerNumberScheduled).To(Equal(int32(2)))
				Expect(updatedThinRuntime.Status.WorkerNumberReady).To(Equal(int32(2)))
				Expect(updatedThinRuntime.Status.WorkerNumberAvailable).To(Equal(int32(2)))

				Expect(updatedThinRuntime.Status.DesiredFuseNumberScheduled).To(Equal(int32(2)))
				Expect(updatedThinRuntime.Status.CurrentFuseNumberScheduled).To(Equal(int32(2)))
				Expect(updatedThinRuntime.Status.FuseNumberReady).To(Equal(int32(2)))

				// Verify other fields
				Expect(updatedThinRuntime.Status.SetupDuration).To(Equal("1m30s"))
				Expect(updatedThinRuntime.Status.Selector).To(Equal("cache=curvine"))
				Expect(len(updatedThinRuntime.Status.Conditions)).To(Equal(2))

				// Verify mounts were set from reference dataset spec
				Expect(len(updatedThinRuntime.Status.Mounts)).To(Equal(1))
				Expect(updatedThinRuntime.Status.Mounts[0].Name).To(Equal("cached-data"))
				Expect(updatedThinRuntime.Status.Mounts[0].MountPoint).To(Equal("dataset://cache-ns/cache-runtime"))
			})

			It("should handle different CacheRuntime phases correctly", func() {
				// Update CacheRuntime to NotReady phase
				cacheRuntimeObj.Status.Master.Phase = datav1alpha1.RuntimePhaseNotReady
				cacheRuntimeObj.Status.Master.Reason = "MasterNotReady"
				cacheRuntimeObj.Status.Worker.Phase = datav1alpha1.RuntimePhaseNotReady
				cacheRuntimeObj.Status.Worker.Reason = "WorkersNotReady"
				cacheRuntimeObj.Status.Client.Phase = datav1alpha1.RuntimePhaseNotReady
				cacheRuntimeObj.Status.Client.Reason = "FusesNotReady"

				err := fakeClient.Status().Update(context.TODO(), &cacheRuntimeObj)
				Expect(err).NotTo(HaveOccurred())

				e := &ReferenceDatasetEngine{
					Client:            fakeClient,
					Log:               fake.NullLogger(),
					name:              refDatasetObj.GetName(),
					namespace:         refDatasetObj.GetNamespace(),
					timeOfLastSync:    time.Now().Add(-defaultSyncRetryDuration),
					syncRetryDuration: defaultSyncRetryDuration,
				}

				ctx := cruntime.ReconcileRequestContext{}
				err = e.Sync(ctx)
				Expect(err).NotTo(HaveOccurred())

				// Verify ThinRuntime status reflects NotReady phases
				updatedThinRuntime := &datav1alpha1.ThinRuntime{}
				err = fakeClient.Get(context.TODO(), types.NamespacedName{
					Namespace: refRuntimeObj.Namespace,
					Name:      refRuntimeObj.Name,
				}, updatedThinRuntime)
				Expect(err).NotTo(HaveOccurred())

				Expect(updatedThinRuntime.Status.MasterPhase).To(Equal(datav1alpha1.RuntimePhaseNotReady))
				Expect(updatedThinRuntime.Status.MasterReason).To(Equal("MasterNotReady"))
				Expect(updatedThinRuntime.Status.WorkerPhase).To(Equal(datav1alpha1.RuntimePhaseNotReady))
				Expect(updatedThinRuntime.Status.WorkerReason).To(Equal("WorkersNotReady"))
				Expect(updatedThinRuntime.Status.FusePhase).To(Equal(datav1alpha1.RuntimePhaseNotReady))
				Expect(updatedThinRuntime.Status.FuseReason).To(Equal("FusesNotReady"))
			})

			It("should handle CacheRuntime with partial readiness", func() {
				// Set partial readiness - master ready, workers not fully ready
				cacheRuntimeObj.Status.Master.Phase = datav1alpha1.RuntimePhaseReady
				cacheRuntimeObj.Status.Worker.Phase = datav1alpha1.RuntimePhaseNotReady
				cacheRuntimeObj.Status.Worker.DesiredReplicas = 3
				cacheRuntimeObj.Status.Worker.CurrentReplicas = 2
				cacheRuntimeObj.Status.Worker.ReadyReplicas = 1
				cacheRuntimeObj.Status.Worker.AvailableReplicas = 1
				cacheRuntimeObj.Status.Worker.UnavailableReplicas = 2
				cacheRuntimeObj.Status.Client.Phase = datav1alpha1.RuntimePhaseReady

				err := fakeClient.Status().Update(context.TODO(), &cacheRuntimeObj)
				Expect(err).NotTo(HaveOccurred())

				e := &ReferenceDatasetEngine{
					Client:            fakeClient,
					Log:               fake.NullLogger(),
					name:              refDatasetObj.GetName(),
					namespace:         refDatasetObj.GetNamespace(),
					timeOfLastSync:    time.Now().Add(-defaultSyncRetryDuration),
					syncRetryDuration: defaultSyncRetryDuration,
				}

				ctx := cruntime.ReconcileRequestContext{}
				err = e.Sync(ctx)
				Expect(err).NotTo(HaveOccurred())

				updatedThinRuntime := &datav1alpha1.ThinRuntime{}
				err = fakeClient.Get(context.TODO(), types.NamespacedName{
					Namespace: refRuntimeObj.Namespace,
					Name:      refRuntimeObj.Name,
				}, updatedThinRuntime)
				Expect(err).NotTo(HaveOccurred())

				// Verify mixed phases
				Expect(updatedThinRuntime.Status.MasterPhase).To(Equal(datav1alpha1.RuntimePhaseReady))
				Expect(updatedThinRuntime.Status.WorkerPhase).To(Equal(datav1alpha1.RuntimePhaseNotReady))
				Expect(updatedThinRuntime.Status.FusePhase).To(Equal(datav1alpha1.RuntimePhaseReady))

				// Verify worker replica counts reflect partial readiness
				Expect(updatedThinRuntime.Status.DesiredWorkerNumberScheduled).To(Equal(int32(3)))
				Expect(updatedThinRuntime.Status.CurrentWorkerNumberScheduled).To(Equal(int32(2)))
				Expect(updatedThinRuntime.Status.WorkerNumberReady).To(Equal(int32(1)))
				Expect(updatedThinRuntime.Status.WorkerNumberAvailable).To(Equal(int32(1)))
				Expect(updatedThinRuntime.Status.WorkerNumberUnavailable).To(Equal(int32(2)))
			})
		})
	})
})
