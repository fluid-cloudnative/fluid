/*
Copyright 2020 The Fluid Authors.

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

package dataset

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func newDatasetScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	_ = datav1alpha1.AddToScheme(s)
	_ = corev1.AddToScheme(s)
	return s
}

func newFakeDatasetClient(objs ...runtime.Object) client.Client {
	return fake.NewFakeClientWithScheme(newDatasetScheme(), objs...)
}

func newTestReconciler(objs ...runtime.Object) *DatasetReconciler {
	c := newFakeDatasetClient(objs...)
	return &DatasetReconciler{
		Client:       c,
		Recorder:     record.NewFakeRecorder(100),
		Log:          zap.New(zap.UseDevMode(true)),
		Scheme:       newDatasetScheme(),
		ResyncPeriod: 5 * time.Second,
	}
}

func newTestReconcilerWithInterceptor(interceptorFuncs interceptor.Funcs, objs ...runtime.Object) *DatasetReconciler {
	s := newDatasetScheme()
	var clientObjs []client.Object
	for _, obj := range objs {
		if co, ok := obj.(client.Object); ok {
			clientObjs = append(clientObjs, co)
		}
	}
	c := ctrlfake.NewClientBuilder().
		WithScheme(s).
		WithRuntimeObjects(objs...).
		WithStatusSubresource(clientObjs...).
		WithInterceptorFuncs(interceptorFuncs).
		Build()
	return &DatasetReconciler{
		Client:       c,
		Recorder:     record.NewFakeRecorder(100),
		Log:          zap.New(zap.UseDevMode(true)),
		Scheme:       s,
		ResyncPeriod: 5 * time.Second,
	}
}

func makeReconcileCtx(r *DatasetReconciler, ds datav1alpha1.Dataset) reconcileRequestContext {
	return reconcileRequestContext{
		Context: context.Background(),
		Log:     r.Log.WithValues("dataset", ds.Namespace+"/"+ds.Name),
		NamespacedName: types.NamespacedName{
			Namespace: ds.Namespace,
			Name:      ds.Name,
		},
		Dataset: ds,
	}
}

var _ = Describe("DatasetReconciler (fake client)", func() {
	Describe("ControllerName", func() {
		It("returns DatasetController", func() {
			r := newTestReconciler()
			Expect(r.ControllerName()).To(Equal("DatasetController"))
			Expect(r.ControllerName()).To(Equal(controllerName))
		})
	})

	Describe("addFinalizerAndRequeue", func() {
		It("adds finalizer and requeues", func() {
			ds := datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{Name: "af", Namespace: "default"},
			}
			r := newTestReconciler(&ds)
			ctx := makeReconcileCtx(r, ds)

			result, err := r.addFinalizerAndRequeue(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{Requeue: true}))

			stored := &datav1alpha1.Dataset{}
			Expect(r.Get(ctx, types.NamespacedName{Namespace: "default", Name: "af"}, stored)).To(Succeed())
			Expect(stored.Finalizers).To(ContainElement(finalizer))
		})
	})

	Describe("reconcileDataset", func() {
		It("returns error for invalid DNS1035 name starting with digit", func() {
			ds := datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{Name: "1invalid", Namespace: "default"},
			}
			r := newTestReconciler(&ds)
			ctx := makeReconcileCtx(r, ds)

			result, err := r.reconcileDataset(ctx, false)
			Expect(err).To(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))
		})

		It("adds finalizer when dataset has none", func() {
			ds := datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{Name: "no-fin", Namespace: "default"},
			}
			r := newTestReconciler(&ds)
			ctx := makeReconcileCtx(r, ds)

			result, err := r.reconcileDataset(ctx, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{Requeue: true}))

			stored := &datav1alpha1.Dataset{}
			Expect(r.Get(ctx, types.NamespacedName{Namespace: "default", Name: "no-fin"}, stored)).To(Succeed())
			Expect(stored.Finalizers).To(ContainElement(finalizer))
		})

		It("advances phase from None to NotBound when finalizer already present", func() {
			ds := datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "none-phase",
					Namespace:  "default",
					Finalizers: []string{finalizer},
				},
				Status: datav1alpha1.DatasetStatus{Phase: datav1alpha1.NoneDatasetPhase},
			}
			r := newTestReconciler(&ds)
			ctx := makeReconcileCtx(r, ds)

			result, err := r.reconcileDataset(ctx, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))
		})

		It("requeues after ResyncPeriod when needRequeue is true", func() {
			ds := datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "needs-requeue",
					Namespace:  "default",
					Finalizers: []string{finalizer},
				},
				Status: datav1alpha1.DatasetStatus{Phase: datav1alpha1.NotBoundDatasetPhase},
			}
			r := newTestReconciler(&ds)
			ctx := makeReconcileCtx(r, ds)

			result, err := r.reconcileDataset(ctx, true)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(r.ResyncPeriod))
		})

		It("delegates to reconcileDatasetDeletion when DeletionTimestamp is set", func() {
			now := metav1.Now()
			ds := datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "with-deletion",
					Namespace:         "default",
					Finalizers:        []string{finalizer},
					DeletionTimestamp: &now,
				},
			}
			r := newTestReconciler(&ds)
			ctx := makeReconcileCtx(r, ds)

			result, err := r.reconcileDataset(ctx, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))
		})

		It("returns error when CheckReferenceDataset fails (multiple dataset mounts)", func() {
			// Two dataset:// mounts → CheckReferenceDataset returns an error
			ds := datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "multi-ref",
					Namespace:  "default",
					Finalizers: []string{finalizer},
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{Name: "m1", MountPoint: "dataset://default/ds1"},
						{Name: "m2", MountPoint: "dataset://default/ds2"},
					},
				},
				Status: datav1alpha1.DatasetStatus{Phase: datav1alpha1.NotBoundDatasetPhase},
			}
			r := newTestReconciler(&ds)
			ctx := makeReconcileCtx(r, ds)

			result, err := r.reconcileDataset(ctx, false)
			Expect(err).To(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))
		})

		It("creates ThinRuntime for reference dataset (single dataset:// mount)", func() {
			// Single dataset:// mount → checkReferenceDataset=true → CreateRuntimeForReferenceDatasetIfNotExist
			ds := datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "ref-ds",
					Namespace:  "default",
					Finalizers: []string{finalizer},
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{Name: "m1", MountPoint: "dataset://default/physical-ds"},
					},
				},
				Status: datav1alpha1.DatasetStatus{Phase: datav1alpha1.NotBoundDatasetPhase},
			}
			r := newTestReconciler(&ds)
			ctx := makeReconcileCtx(r, ds)

			result, err := r.reconcileDataset(ctx, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))
		})

		It("returns error when CreateRuntimeForReferenceDatasetIfNotExist fails", func() {
			ds := datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "ref-ds-create-err",
					Namespace:  "default",
					Finalizers: []string{finalizer},
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{Name: "m1", MountPoint: "dataset://default/physical-ds"},
					},
				},
				Status: datav1alpha1.DatasetStatus{Phase: datav1alpha1.NotBoundDatasetPhase},
			}
			createErr := fmt.Errorf("injected create error")
			r := newTestReconcilerWithInterceptor(interceptor.Funcs{
				Create: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
					return createErr
				},
			}, &ds)
			ctx := makeReconcileCtx(r, ds)

			result, err := r.reconcileDataset(ctx, false)
			Expect(err).To(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))
		})
	})

	Describe("reconcileDatasetDeletion", func() {
		It("removes finalizer when no pods block and no DatasetRef", func() {
			now := metav1.Now()
			ds := datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "del-clean",
					Namespace:         "default",
					Finalizers:        []string{finalizer},
					DeletionTimestamp: &now,
				},
			}
			r := newTestReconciler(&ds)
			ctx := makeReconcileCtx(r, ds)

			result, err := r.reconcileDatasetDeletion(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))

			stored := &datav1alpha1.Dataset{}
			getErr := r.Get(ctx, types.NamespacedName{Namespace: "default", Name: "del-clean"}, stored)
			Expect(apierrors.IsNotFound(getErr)).To(BeTrue())
		})

		It("requeues when DatasetRef still has live entries", func() {
			now := metav1.Now()
			// Create both the main dataset and the referenced dataset in the fake store.
			// With the reference alive, deletion is blocked.
			ref := datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{Name: "ref-dataset", Namespace: "default"},
			}
			ds := datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "del-with-ref",
					Namespace:         "default",
					Finalizers:        []string{finalizer},
					DeletionTimestamp: &now,
				},
				Status: datav1alpha1.DatasetStatus{
					DatasetRef: []string{"default/ref-dataset"},
				},
			}
			r := newTestReconciler(&ref, &ds)
			ctx := makeReconcileCtx(r, ds)

			result, err := r.reconcileDatasetDeletion(ctx)
			Expect(err).NotTo(HaveOccurred())
			// Blocked by live reference – must requeue
			Expect(result.RequeueAfter > 0 || result.Requeue).To(BeTrue())
		})

		It("updates status and requeues when stale DatasetRef entries are pruned", func() {
			now := metav1.Now()
			// Only ref-alive exists; ref-gone does NOT → RemoveNotFoundDatasetRef prunes it.
			// Pruned list differs from original → status update + 1s requeue.
			refAlive := datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{Name: "ref-alive", Namespace: "default"},
			}
			ds := datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "del-prune-ref",
					Namespace:         "default",
					Finalizers:        []string{finalizer},
					DeletionTimestamp: &now,
				},
				Status: datav1alpha1.DatasetStatus{
					DatasetRef: []string{"default/ref-alive", "default/ref-gone"},
				},
			}
			r := newTestReconciler(&refAlive, &ds)
			ctx := makeReconcileCtx(r, ds)

			result, err := r.reconcileDatasetDeletion(ctx)
			Expect(err).NotTo(HaveOccurred())
			// After pruning, ref-alive is still live → blocked → requeue
			Expect(result.RequeueAfter > 0 || result.Requeue).To(BeTrue())
		})

		It("requeues when a PVC blocks dataset deletion", func() {
			const blockedDatasetName = "del-blocked"

			now := metav1.Now()
			// Create a PVC with the Fluid annotation in the same namespace/name as the dataset.
			pvc := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      blockedDatasetName,
					Namespace: "default",
					Annotations: map[string]string{
						"CreatedBy": "fluid",
					},
				},
			}
			// Create a running pod that mounts the PVC.
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "using-pod",
					Namespace: "default",
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "data",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: blockedDatasetName,
								},
							},
						},
					},
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
				},
			}
			ds := datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:              blockedDatasetName,
					Namespace:         "default",
					Finalizers:        []string{finalizer},
					DeletionTimestamp: &now,
				},
			}
			r := newTestReconciler(&ds, pvc, pod)
			ctx := makeReconcileCtx(r, ds)

			result, err := r.reconcileDatasetDeletion(ctx)
			// ShouldDeleteDataset returns error (pod is still using the PVC) → requeue
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter > 0 || result.Requeue).To(BeTrue())
		})

		It("returns error when r.Update fails while removing finalizer", func() {
			now := metav1.Now()
			ds := datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "del-update-err",
					Namespace:         "default",
					Finalizers:        []string{finalizer},
					DeletionTimestamp: &now,
				},
			}
			updateErr := fmt.Errorf("injected update error")
			r := newTestReconcilerWithInterceptor(interceptor.Funcs{
				Update: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
					return updateErr
				},
			}, &ds)
			ctx := makeReconcileCtx(r, ds)

			result, err := r.reconcileDatasetDeletion(ctx)
			Expect(err).To(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))
		})
	})

	Describe("addFinalizerAndRequeue error path", func() {
		It("returns error when Update fails", func() {
			ds := datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{Name: "af-err", Namespace: "default"},
			}
			updateErr := fmt.Errorf("injected update error")
			r := newTestReconcilerWithInterceptor(interceptor.Funcs{
				Update: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
					return updateErr
				},
			}, &ds)
			ctx := makeReconcileCtx(r, ds)

			result, err := r.addFinalizerAndRequeue(ctx)
			Expect(err).To(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))
		})
	})

	Describe("reconcileDataset status update error path", func() {
		It("returns error when status update fails for NoneDatasetPhase", func() {
			ds := datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "status-err",
					Namespace:  "default",
					Finalizers: []string{finalizer},
				},
				Status: datav1alpha1.DatasetStatus{Phase: datav1alpha1.NoneDatasetPhase},
			}
			statusErr := fmt.Errorf("injected status update error")
			r := newTestReconcilerWithInterceptor(interceptor.Funcs{
				SubResourceUpdate: func(ctx context.Context, c client.Client, subResourceName string, obj client.Object, opts ...client.SubResourceUpdateOption) error {
					return statusErr
				},
			}, &ds)
			ctx := makeReconcileCtx(r, ds)

			result, err := r.reconcileDataset(ctx, false)
			Expect(err).To(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))
		})
	})

	Describe("reconcileDatasetDeletion error paths", func() {
		It("requeues when status update fails after datasetRef is pruned", func() {
			now := metav1.Now()
			// Only ref-gone listed → will be pruned → triggers status update → inject error
			ds := datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "del-status-err",
					Namespace:         "default",
					Finalizers:        []string{finalizer},
					DeletionTimestamp: &now,
				},
				Status: datav1alpha1.DatasetStatus{
					DatasetRef: []string{"default/ref-gone"},
				},
			}
			statusErr := fmt.Errorf("injected status update error")
			r := newTestReconcilerWithInterceptor(interceptor.Funcs{
				SubResourceUpdate: func(ctx context.Context, c client.Client, subResourceName string, obj client.Object, opts ...client.SubResourceUpdateOption) error {
					return statusErr
				},
			}, &ds)
			ctx := makeReconcileCtx(r, ds)

			result, err := r.reconcileDatasetDeletion(ctx)
			// Status update failed → returns RequeueAfterInterval(10s)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter > 0 || result.Requeue).To(BeTrue())
		})
	})
})
