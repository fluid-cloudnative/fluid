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
	"errors"
	"os"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	cdataload "github.com/fluid-cloudnative/fluid/pkg/dataload"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	fluiderrors "github.com/fluid-cloudnative/fluid/pkg/errors"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

// fakeOperation is a minimal dataoperation.OperationInterface for routing tests.
type fakeOperation struct {
	operationType dataoperation.OperationType
	object        client.Object
}

var _ dataoperation.OperationInterface = (*fakeOperation)(nil)

func (f *fakeOperation) HasPrecedingOperation() bool { return false }

func (f *fakeOperation) GetOperationObject() client.Object { return f.object }

func (f *fakeOperation) GetPossibleTargetDatasetNamespacedNames() []types.NamespacedName {
	return nil
}

// GetTargetDataset is not used by the cache engine generators, which look the
// dataset up themselves, so fail loudly if a test starts relying on it.
func (f *fakeOperation) GetTargetDataset() (*datav1alpha1.Dataset, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeOperation) GetReleaseNameSpacedName() types.NamespacedName {
	return types.NamespacedName{}
}

func (f *fakeOperation) GetChartsDirectory() string { return "" }

func (f *fakeOperation) GetOperationType() dataoperation.OperationType { return f.operationType }

func (f *fakeOperation) UpdateOperationApiStatus(*datav1alpha1.OperationStatus) error { return nil }

func (f *fakeOperation) Validate(cruntime.ReconcileRequestContext) ([]datav1alpha1.Condition, error) {
	return nil, nil
}

func (f *fakeOperation) UpdateStatusInfoForCompleted(map[string]string) error { return nil }

func (f *fakeOperation) SetTargetDatasetStatusInProgress(*datav1alpha1.Dataset) {}

func (f *fakeOperation) RemoveTargetDatasetStatusInProgress(*datav1alpha1.Dataset) {}

func (f *fakeOperation) GetStatusHandler() dataoperation.StatusHandler { return nil }

func (f *fakeOperation) GetTTL() (*int32, error) { return nil, nil }

func (f *fakeOperation) GetParallelTaskNumber() int32 { return 1 }

var _ = Describe("CacheEngine data operations", Label("pkg.ddc.cache.engine.data_operation_test.go"), func() {
	Describe("CheckRuntimeReady", func() {
		newEngine := func(masterDisabled bool, masterPhase, workerPhase datav1alpha1.RuntimePhase) *CacheEngine {
			runtimeObj := &datav1alpha1.CacheRuntime{
				ObjectMeta: metav1.ObjectMeta{Name: "test-runtime", Namespace: "default"},
			}
			runtimeObj.Spec.Master.Disabled = masterDisabled
			runtimeObj.Status.Master.Phase = masterPhase
			runtimeObj.Status.Worker.Phase = workerPhase
			return &CacheEngine{
				name:      "test-runtime",
				namespace: "default",
				Client:    fake.NewFakeClientWithScheme(CacheEngineTestScheme, runtimeObj),
				Log:       ctrl.Log.WithName("test"),
			}
		}

		It("returns false when the runtime does not exist", func() {
			engine := &CacheEngine{
				name:      "missing",
				namespace: "default",
				Client:    fake.NewFakeClientWithScheme(CacheEngineTestScheme),
				Log:       ctrl.Log.WithName("test"),
			}
			Expect(engine.CheckRuntimeReady()).To(BeFalse())
		})

		DescribeTable("checks the master and worker phases",
			func(masterDisabled bool, masterPhase, workerPhase datav1alpha1.RuntimePhase, want bool) {
				Expect(newEngine(masterDisabled, masterPhase, workerPhase).CheckRuntimeReady()).To(Equal(want))
			},
			Entry("master-worker, both ready", false, datav1alpha1.RuntimePhaseReady, datav1alpha1.RuntimePhaseReady, true),
			Entry("master-worker, worker partially ready", false, datav1alpha1.RuntimePhaseReady, datav1alpha1.RuntimePhasePartialReady, true),
			Entry("master-worker, worker not ready", false, datav1alpha1.RuntimePhaseReady, datav1alpha1.RuntimePhaseNotReady, false),
			Entry("master-worker, master not ready", false, datav1alpha1.RuntimePhaseNotReady, datav1alpha1.RuntimePhaseReady, false),
			Entry("master-worker, master only partially ready", false, datav1alpha1.RuntimePhasePartialReady, datav1alpha1.RuntimePhaseReady, false),
			Entry("master-worker, nothing set up yet", false, datav1alpha1.RuntimePhaseNone, datav1alpha1.RuntimePhaseNone, false),
			Entry("master-worker, worker not set up yet", false, datav1alpha1.RuntimePhaseReady, datav1alpha1.RuntimePhaseNone, false),
			Entry("master disabled, worker ready", true, datav1alpha1.RuntimePhaseNone, datav1alpha1.RuntimePhaseReady, true),
			Entry("master disabled, worker partially ready", true, datav1alpha1.RuntimePhaseNone, datav1alpha1.RuntimePhasePartialReady, true),
			Entry("master disabled, worker not ready", true, datav1alpha1.RuntimePhaseNone, datav1alpha1.RuntimePhaseNotReady, false),
			Entry("master disabled, worker not set up yet", true, datav1alpha1.RuntimePhaseNone, datav1alpha1.RuntimePhaseNone, false),
		)
	})

	Describe("GetDataOperationValueFile", func() {
		var engine *CacheEngine

		BeforeEach(func() {
			engine = &CacheEngine{
				name:      "test-runtime",
				namespace: "default",
				Client:    fake.NewFakeClientWithScheme(CacheEngineTestScheme),
				Log:       ctrl.Log.WithName("test"),
			}
		})

		withKind := func(obj client.Object, kind string) client.Object {
			obj.GetObjectKind().SetGroupVersionKind(datav1alpha1.GroupVersion.WithKind(kind))
			return obj
		}

		DescribeTable("returns not supported for operations the cache runtime does not implement",
			func(opType dataoperation.OperationType, object client.Object) {
				valueFile, err := engine.GetDataOperationValueFile(cruntime.ReconcileRequestContext{}, &fakeOperation{operationType: opType, object: object})
				Expect(err).To(HaveOccurred())
				Expect(fluiderrors.IsNotSupported(err)).To(BeTrue())
				Expect(valueFile).To(BeEmpty())
			},
			Entry("DataBackup", dataoperation.DataBackupType, withKind(&datav1alpha1.DataBackup{}, "DataBackup")),
			Entry("DataMigrate", dataoperation.DataMigrateType, withKind(&datav1alpha1.DataMigrate{}, "DataMigrate")),
			Entry("unknown operation type", dataoperation.OperationType("DataUnknown"), withKind(&datav1alpha1.DataLoad{}, "DataLoad")),
		)

		It("routes DataLoad to the DataLoad value file generator", func() {
			// a mismatched object makes the generator fail early with its own message
			_, err := engine.GetDataOperationValueFile(cruntime.ReconcileRequestContext{},
				&fakeOperation{operationType: dataoperation.DataLoadType, object: &datav1alpha1.DataProcess{}})
			Expect(err).To(MatchError(ContainSubstring("is not a DataLoad")))
		})

		It("writes a DataLoad value file built from the runtime class DataOperationSpecs", func() {
			runtimeObj := &datav1alpha1.CacheRuntime{
				ObjectMeta: metav1.ObjectMeta{Name: "test-runtime", Namespace: "default"},
				Spec:       datav1alpha1.CacheRuntimeSpec{RuntimeClassName: "test-class"},
			}
			runtimeClass := &datav1alpha1.CacheRuntimeClass{
				ObjectMeta: metav1.ObjectMeta{Name: "test-class"},
				Topology:   &datav1alpha1.RuntimeTopology{Worker: &datav1alpha1.RuntimeComponentDefinition{}},
				DataOperationSpecs: []datav1alpha1.DataOperationSpec{{
					Name:    string(dataoperation.DataLoadType),
					Image:   "dataload:v1",
					Command: []string{"/bin/sh", "-c", "load.sh"},
				}},
			}
			dataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{Name: "test-dataset", Namespace: "default", UID: "dataset-uid"},
			}
			dataload := &datav1alpha1.DataLoad{
				ObjectMeta: metav1.ObjectMeta{Name: "test-load", Namespace: "default"},
				Spec: datav1alpha1.DataLoadSpec{
					Dataset: datav1alpha1.TargetDataset{Name: "test-dataset", Namespace: "default"},
					Target:  []datav1alpha1.TargetPath{{Path: "/data", Replicas: 2}},
				},
			}
			c := fake.NewFakeClientWithScheme(CacheEngineTestScheme, runtimeObj, runtimeClass, dataset, dataload)
			engine.Client = c
			_, err := engine.getBoundRuntimeClass(context.Background(), runtimeObj)
			Expect(err).NotTo(HaveOccurred())

			valueFile, err := engine.GetDataOperationValueFile(cruntime.ReconcileRequestContext{Client: c},
				&fakeOperation{operationType: dataoperation.DataLoadType, object: dataload})
			Expect(err).NotTo(HaveOccurred())
			// best effort: the generator never closes its temp file handle, which blocks removal on Windows
			DeferCleanup(func() { _ = os.Remove(valueFile) })
			Expect(valueFile).To(ContainSubstring("default-test-load-loader-values.yaml"))

			data, err := os.ReadFile(valueFile)
			Expect(err).NotTo(HaveOccurred())
			var value cdataload.DataLoadValue
			Expect(yaml.Unmarshal(data, &value)).To(Succeed())

			Expect(value.Name).To(Equal("test-load"))
			// short namespaced names are used as-is; the dataset UID only kicks in past the DNS label limit
			Expect(value.OwnerDatasetId).To(Equal("default-test-dataset"))
			info := value.DataLoadInfo
			Expect(info.Image).To(Equal("dataload:v1"))
			Expect(info.Command).To(Equal([]string{"/bin/sh", "-c", "load.sh"}))
			Expect(info.TargetDataset).To(Equal("test-dataset"))
			Expect(info.TargetPaths).To(Equal([]cdataload.TargetPath{{Path: "/data", Replicas: 2}}))
			Expect(info.Envs).To(ConsistOf(cdataload.Env{Name: "FLUID_RUNTIME_CONFIG_PATH", Value: "/etc/fluid/config/runtime.json"}))

			Expect(info.Volumes).To(HaveLen(1))
			Expect(info.Volumes[0].Name).To(Equal("fluid-runtime-test-runtime-config"))
			Expect(info.Volumes[0].ConfigMap).NotTo(BeNil())
			Expect(info.Volumes[0].ConfigMap.Name).To(Equal("fluid-runtime-config-test-runtime"))
			Expect(info.VolumeMounts).To(HaveLen(1))
			Expect(info.VolumeMounts[0].Name).To(Equal("fluid-runtime-test-runtime-config"))
			Expect(info.VolumeMounts[0].MountPath).To(Equal("/etc/fluid/config"))
		})

		It("routes DataProcess to the DataProcess value file generator", func() {
			_, err := engine.GetDataOperationValueFile(cruntime.ReconcileRequestContext{},
				&fakeOperation{operationType: dataoperation.DataProcessType, object: &datav1alpha1.DataLoad{}})
			Expect(err).To(MatchError(ContainSubstring("is not of type DataProcess")))
		})
	})

	Describe("findDataOperationSpec", func() {
		specs := []datav1alpha1.DataOperationSpec{
			{Name: string(dataoperation.DataLoadType), Image: "dataload:v1"},
			{Name: string(dataoperation.DataProcessType), Image: "dataprocess:v1"},
		}

		It("returns the spec with the matching name", func() {
			spec := findDataOperationSpec(specs, dataoperation.DataProcessType)
			Expect(spec).NotTo(BeNil())
			Expect(spec.Image).To(Equal("dataprocess:v1"))
		})

		It("returns nil when no spec matches", func() {
			Expect(findDataOperationSpec(specs, dataoperation.DataBackupType)).To(BeNil())
		})

		It("returns nil for an empty list", func() {
			Expect(findDataOperationSpec(nil, dataoperation.DataLoadType)).To(BeNil())
		})
	})
})
