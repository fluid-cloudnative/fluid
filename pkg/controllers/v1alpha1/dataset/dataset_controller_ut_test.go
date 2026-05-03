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

package dataset

import (
	"context"
	"fmt"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/controllers/deploy"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Dataset Controller Unit", func() {
	var scheme *runtime.Scheme
	var scaleoutPatch *gomonkey.Patches

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		Expect(datav1alpha1.AddToScheme(scheme)).NotTo(HaveOccurred())
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		scaleoutPatch = gomonkey.ApplyFunc(deploy.ScaleoutRuntimeControllerOnDemand,
			func(client.Client, types.NamespacedName, logr.Logger) (string, bool, error) {
				return "", false, nil
			})
	})

	AfterEach(func() {
		scaleoutPatch.Reset()
	})

	Describe("ControllerName", func() {
		It("should return the dataset controller name", func() {
			r := &DatasetReconciler{}

			Expect(r.ControllerName()).To(Equal(controllerName))
		})
	})

	Describe("Reconcile", func() {
		It("should return empty result when dataset is not found", func() {
			r := &DatasetReconciler{
				Client:   fake.NewFakeClientWithScheme(scheme),
				Scheme:   scheme,
				Recorder: record.NewFakeRecorder(16),
			}

			result, err := r.Reconcile(context.TODO(), ctrl.Request{
				NamespacedName: types.NamespacedName{Name: "missing", Namespace: "default"},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))
		})

		It("should return an error when dataset name violates DNS-1035 validation", func() {
			invalidName := "20-dataset"
			dataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      invalidName,
					Namespace: "default",
				},
			}

			r := &DatasetReconciler{
				Client:   fake.NewFakeClientWithScheme(scheme, dataset),
				Scheme:   scheme,
				Recorder: record.NewFakeRecorder(16),
			}

			_, err := r.Reconcile(context.TODO(), ctrl.Request{
				NamespacedName: types.NamespacedName{Name: invalidName, Namespace: "default"},
			})

			Expect(err).To(HaveOccurred())
		})

		It("should add the finalizer and requeue for a new dataset", func() {
			dataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "sample-dataset",
					Namespace:  "default",
					Generation: 1,
				},
			}

			r := &DatasetReconciler{
				Client:       fake.NewFakeClientWithScheme(scheme, dataset),
				Scheme:       scheme,
				Recorder:     record.NewFakeRecorder(16),
				ResyncPeriod: 30 * time.Second,
			}

			result, err := r.Reconcile(context.TODO(), ctrl.Request{
				NamespacedName: types.NamespacedName{Name: dataset.Name, Namespace: dataset.Namespace},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(result.Requeue).To(BeTrue())

			updatedDataset := &datav1alpha1.Dataset{}
			Expect(r.Get(context.TODO(), types.NamespacedName{Name: dataset.Name, Namespace: dataset.Namespace}, updatedDataset)).To(Succeed())
			Expect(updatedDataset.GetFinalizers()).To(ContainElement(finalizer))
		})

		It("should update an initialized dataset phase to not bound without requeue", func() {
			dataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "validated-dataset",
					Namespace:  "default",
					Generation: 1,
					Finalizers: []string{finalizer},
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{{
						MountPoint: "oss://bucket/path",
						Name:       "main",
					}},
				},
			}

			r := &DatasetReconciler{
				Client:       fake.NewFakeClientWithScheme(scheme, dataset),
				Scheme:       scheme,
				Recorder:     record.NewFakeRecorder(16),
				ResyncPeriod: 30 * time.Second,
			}

			result, err := r.Reconcile(context.TODO(), ctrl.Request{
				NamespacedName: types.NamespacedName{Name: dataset.Name, Namespace: dataset.Namespace},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))

			updatedDataset := &datav1alpha1.Dataset{}
			Expect(r.Get(context.TODO(), types.NamespacedName{Name: dataset.Name, Namespace: dataset.Namespace}, updatedDataset)).To(Succeed())
			Expect(updatedDataset.Status.Phase).To(Equal(datav1alpha1.NotBoundDatasetPhase))
			Expect(updatedDataset.Status.Conditions).To(BeEmpty())
		})

		It("should requeue after the resync period when runtime scaleout requested a retry", func() {
			dataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "requeue-dataset",
					Namespace:  "default",
					Generation: 1,
					Finalizers: []string{finalizer},
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{{
						MountPoint: "oss://bucket/requeue",
						Name:       "main",
					}},
				},
				Status: datav1alpha1.DatasetStatus{
					Phase: datav1alpha1.NotBoundDatasetPhase,
				},
			}

			r := &DatasetReconciler{
				Client:       fake.NewFakeClientWithScheme(scheme, dataset),
				Scheme:       scheme,
				Recorder:     record.NewFakeRecorder(16),
				ResyncPeriod: 45 * time.Second,
			}

			ctx := reconcileRequestContext{
				Context: context.TODO(),
				Dataset: *dataset,
				NamespacedName: types.NamespacedName{
					Name:      dataset.Name,
					Namespace: dataset.Namespace,
				},
			}

			result, err := r.reconcileDataset(ctx, true)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(45 * time.Second))
		})

		It("should requeue after the resync period when runtime scaleout returns an error", func() {
			scaleoutPatch.Reset()
			scaleoutPatch = gomonkey.ApplyFunc(deploy.ScaleoutRuntimeControllerOnDemand,
				func(client.Client, types.NamespacedName, logr.Logger) (string, bool, error) {
					return "", false, fmt.Errorf("scaleout failed")
				})

			dataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "scaleout-requeue-dataset",
					Namespace:  "default",
					Generation: 1,
					Finalizers: []string{finalizer},
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{{
						MountPoint: "oss://bucket/requeue-after-scaleout-error",
						Name:       "main",
					}},
				},
				Status: datav1alpha1.DatasetStatus{
					Phase: datav1alpha1.NotBoundDatasetPhase,
				},
			}

			r := &DatasetReconciler{
				Client:       fake.NewFakeClientWithScheme(scheme, dataset),
				Scheme:       scheme,
				Recorder:     record.NewFakeRecorder(16),
				ResyncPeriod: 20 * time.Second,
			}

			result, err := r.Reconcile(context.TODO(), ctrl.Request{
				NamespacedName: types.NamespacedName{Name: dataset.Name, Namespace: dataset.Namespace},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(20 * time.Second))
		})
	})

	Describe("addFinalizerAndRequeue", func() {
		It("should add the controller finalizer and requeue immediately", func() {
			storedDataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "finalizer-dataset",
					Namespace:  "default",
					Generation: 3,
				},
			}

			r := &DatasetReconciler{
				Client:   fake.NewFakeClientWithScheme(scheme, storedDataset),
				Scheme:   scheme,
				Recorder: record.NewFakeRecorder(16),
			}

			currentDataset := datav1alpha1.Dataset{}
			Expect(r.Get(context.TODO(), types.NamespacedName{Name: storedDataset.Name, Namespace: storedDataset.Namespace}, &currentDataset)).To(Succeed())

			result, err := r.addFinalizerAndRequeue(reconcileRequestContext{
				Context: context.TODO(),
				Dataset: currentDataset,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(result.Requeue).To(BeTrue())

			updatedDataset := &datav1alpha1.Dataset{}
			Expect(r.Get(context.TODO(), types.NamespacedName{Name: storedDataset.Name, Namespace: storedDataset.Namespace}, updatedDataset)).To(Succeed())
			Expect(updatedDataset.GetFinalizers()).To(ContainElement(finalizer))
		})
	})

	Describe("reconcileDatasetDeletion", func() {
		It("should remove the controller finalizer when the dataset can be deleted", func() {
			deletionTime := metav1.NewTime(time.Now())
			storedDataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "deleting-dataset",
					Namespace:         "default",
					Finalizers:        []string{finalizer},
					DeletionTimestamp: &deletionTime,
				},
			}

			r := &DatasetReconciler{
				Client:   fake.NewFakeClientWithScheme(scheme, storedDataset),
				Scheme:   scheme,
				Recorder: record.NewFakeRecorder(16),
			}

			currentDataset := datav1alpha1.Dataset{}
			Expect(r.Get(context.TODO(), types.NamespacedName{Name: storedDataset.Name, Namespace: storedDataset.Namespace}, &currentDataset)).To(Succeed())

			result, err := r.reconcileDatasetDeletion(reconcileRequestContext{
				Context: context.TODO(),
				Dataset: currentDataset,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))

			updatedDataset := &datav1alpha1.Dataset{}
			err = r.Get(context.TODO(), types.NamespacedName{Name: storedDataset.Name, Namespace: storedDataset.Namespace}, updatedDataset)
			if err == nil {
				Expect(updatedDataset.GetFinalizers()).NotTo(ContainElement(finalizer))
			} else {
				Expect(client.IgnoreNotFound(err)).To(BeNil())
			}
		})

		It("should requeue when referenced datasets are still mounted", func() {
			deletionTime := metav1.NewTime(time.Now())
			dataset := datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "referenced-dataset",
					Namespace:         "default",
					Finalizers:        []string{finalizer},
					DeletionTimestamp: &deletionTime,
				},
				Status: datav1alpha1.DatasetStatus{
					DatasetRef: []string{"consumer-a"},
				},
			}

			r := &DatasetReconciler{
				Client:   fake.NewFakeClientWithScheme(scheme, dataset.DeepCopy()),
				Scheme:   scheme,
				Recorder: record.NewFakeRecorder(16),
			}

			result, err := r.reconcileDatasetDeletion(reconcileRequestContext{
				Context: context.TODO(),
				Dataset: dataset,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(10 * time.Second))
		})

		It("should prune missing dataset references and requeue immediately", func() {
			deletionTime := metav1.NewTime(time.Now())
			storedDataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "dataset-ref-prune",
					Namespace:         "default",
					Finalizers:        []string{finalizer},
					DeletionTimestamp: &deletionTime,
				},
				Status: datav1alpha1.DatasetStatus{
					DatasetRef: []string{"default/existing-ref", "default/missing-ref"},
				},
			}
			existingRef := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "existing-ref",
					Namespace: "default",
				},
			}

			r := &DatasetReconciler{
				Client:   fake.NewFakeClientWithScheme(scheme, storedDataset, existingRef),
				Scheme:   scheme,
				Recorder: record.NewFakeRecorder(16),
			}

			currentDataset := datav1alpha1.Dataset{}
			Expect(r.Get(context.TODO(), types.NamespacedName{Name: storedDataset.Name, Namespace: storedDataset.Namespace}, &currentDataset)).To(Succeed())

			result, err := r.reconcileDatasetDeletion(reconcileRequestContext{
				Context: context.TODO(),
				Dataset: currentDataset,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(1 * time.Second))

			updatedDataset := &datav1alpha1.Dataset{}
			Expect(r.Get(context.TODO(), types.NamespacedName{Name: storedDataset.Name, Namespace: storedDataset.Namespace}, updatedDataset)).To(Succeed())
			Expect(updatedDataset.Status.DatasetRef).To(Equal([]string{"default/existing-ref"}))
		})
	})

	Describe("reconcileDataset", func() {
		It("should return an error when a reference dataset mixes dataset and non-dataset mounts", func() {
			dataset := datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "invalid-reference-dataset",
					Namespace:  "default",
					Generation: 1,
					Finalizers: []string{finalizer},
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{
							MountPoint: "dataset://fluid/source-dataset",
							Name:       "reference",
						},
						{
							MountPoint: "oss://bucket/extra",
							Name:       "extra",
						},
					},
				},
				Status: datav1alpha1.DatasetStatus{
					Phase: datav1alpha1.NotBoundDatasetPhase,
				},
			}

			r := &DatasetReconciler{
				Client:   fake.NewFakeClientWithScheme(scheme, dataset.DeepCopy()),
				Scheme:   scheme,
				Recorder: record.NewFakeRecorder(16),
			}

			_, err := r.reconcileDataset(reconcileRequestContext{
				Context: context.TODO(),
				Dataset: dataset,
				NamespacedName: types.NamespacedName{
					Name:      dataset.Name,
					Namespace: dataset.Namespace,
				},
			}, false)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not validated"))
		})

		It("should create a thin runtime for a reference dataset and keep owner metadata", func() {
			dataset := datav1alpha1.Dataset{
				TypeMeta: metav1.TypeMeta{
					Kind:       datav1alpha1.Datasetkind,
					APIVersion: datav1alpha1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:       "reference-dataset",
					Namespace:  "default",
					UID:        types.UID("dataset-uid"),
					Generation: 1,
					Finalizers: []string{finalizer},
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{{
						MountPoint: "dataset://fluid/source-dataset",
						Name:       "reference",
					}},
				},
				Status: datav1alpha1.DatasetStatus{
					Phase: datav1alpha1.NotBoundDatasetPhase,
				},
			}

			r := &DatasetReconciler{
				Client:   fake.NewFakeClientWithScheme(scheme, dataset.DeepCopy()),
				Scheme:   scheme,
				Recorder: record.NewFakeRecorder(16),
			}

			result, err := r.reconcileDataset(reconcileRequestContext{
				Context: context.TODO(),
				Dataset: dataset,
				NamespacedName: types.NamespacedName{
					Name:      dataset.Name,
					Namespace: dataset.Namespace,
				},
			}, false)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))

			thinRuntime := &datav1alpha1.ThinRuntime{}
			Expect(r.Get(context.TODO(), types.NamespacedName{Name: dataset.Name, Namespace: dataset.Namespace}, thinRuntime)).To(Succeed())
			Expect(thinRuntime.OwnerReferences).To(HaveLen(1))
			Expect(thinRuntime.OwnerReferences[0].Name).To(Equal(dataset.Name))
			Expect(thinRuntime.OwnerReferences[0].UID).To(Equal(dataset.UID))
			Expect(ptr.Deref(thinRuntime.OwnerReferences[0].Controller, false)).To(BeTrue())
		})
	})
})
