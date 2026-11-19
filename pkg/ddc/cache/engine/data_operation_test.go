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

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	fluiderrors "github.com/fluid-cloudnative/fluid/pkg/errors"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("CacheEngine data operation helpers", func() {
	Describe("CheckRuntimeReady", func() {
		DescribeTable("checks the readiness for the runtime topology",
			func(disableMaster bool, masterPhase, workerPhase datav1alpha1.RuntimePhase, expected bool) {
				runtime := &datav1alpha1.CacheRuntime{
					ObjectMeta: metav1.ObjectMeta{Name: "runtime", Namespace: "default"},
					Spec: datav1alpha1.CacheRuntimeSpec{
						Master: datav1alpha1.CacheRuntimeMasterSpec{
							RuntimeComponentCommonSpec: datav1alpha1.RuntimeComponentCommonSpec{Disabled: disableMaster},
						},
					},
					Status: datav1alpha1.CacheRuntimeStatus{
						RuntimeComponentStatusCollection: datav1alpha1.RuntimeComponentStatusCollection{
							Master: datav1alpha1.RuntimeComponentStatus{Phase: masterPhase},
							Worker: datav1alpha1.RuntimeComponentStatus{Phase: workerPhase},
						},
					},
				}
				engine := &CacheEngine{
					Client:    fake.NewFakeClientWithScheme(CacheEngineTestScheme, runtime),
					name:      runtime.Name,
					namespace: runtime.Namespace,
				}

				Expect(engine.CheckRuntimeReady()).To(Equal(expected))
			},
			Entry("master-disabled worker ready", true, datav1alpha1.RuntimePhaseNotReady, datav1alpha1.RuntimePhaseReady, true),
			Entry("master-disabled worker partially ready", true, datav1alpha1.RuntimePhaseNotReady, datav1alpha1.RuntimePhasePartialReady, true),
			Entry("master-disabled worker not ready", true, datav1alpha1.RuntimePhaseReady, datav1alpha1.RuntimePhaseNotReady, false),
			Entry("master-worker ready", false, datav1alpha1.RuntimePhaseReady, datav1alpha1.RuntimePhaseReady, true),
			Entry("master-worker partially ready", false, datav1alpha1.RuntimePhaseReady, datav1alpha1.RuntimePhasePartialReady, true),
			Entry("master-worker master not ready", false, datav1alpha1.RuntimePhaseNotReady, datav1alpha1.RuntimePhaseReady, false),
			Entry("master-worker worker not ready", false, datav1alpha1.RuntimePhaseReady, datav1alpha1.RuntimePhaseNotReady, false),
		)

		It("returns false when the runtime cannot be fetched", func() {
			engine := &CacheEngine{
				Client:    fake.NewFakeClientWithScheme(CacheEngineTestScheme),
				name:      "missing",
				namespace: "default",
			}

			Expect(engine.CheckRuntimeReady()).To(BeFalse())
		})
	})

	Describe("GetDataOperationValueFile", func() {
		It("routes DataLoad operations to the DataLoad generator", func() {
			operation := testValueFileOperation{
				operationType: dataoperation.DataLoadType,
				object:        &datav1alpha1.DataProcess{},
			}
			engine := &CacheEngine{name: "runtime"}

			_, err := engine.GetDataOperationValueFile(cruntime.ReconcileRequestContext{Context: context.Background()}, operation)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("is not a DataLoad"))
		})

		It("routes DataProcess operations to the DataProcess generator", func() {
			operation := testValueFileOperation{
				operationType: dataoperation.DataProcessType,
				object:        &datav1alpha1.DataLoad{},
			}
			engine := &CacheEngine{name: "runtime"}

			_, err := engine.GetDataOperationValueFile(cruntime.ReconcileRequestContext{Context: context.Background()}, operation)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("is not of type DataProcess"))
		})

		DescribeTable("returns NotSupported for unimplemented operation generators",
			func(operationType dataoperation.OperationType) {
				operation := testValueFileOperation{
					operationType: operationType,
					object: &datav1alpha1.DataLoad{
						TypeMeta: metav1.TypeMeta{APIVersion: "data.fluid.io/v1alpha1", Kind: "DataLoad"},
					},
				}
				engine := &CacheEngine{name: "runtime"}

				_, err := engine.GetDataOperationValueFile(cruntime.ReconcileRequestContext{Context: context.Background()}, operation)

				Expect(fluiderrors.IsNotSupported(err)).To(BeTrue())
			},
			Entry("DataMigrate", dataoperation.DataMigrateType),
			Entry("DataBackup", dataoperation.DataBackupType),
			Entry("unknown operation", dataoperation.OperationType("Unknown")),
		)

		It("includes the operation object's resource in the NotSupported error", func() {
			object := &datav1alpha1.DataLoad{
				TypeMeta: metav1.TypeMeta{APIVersion: "data.fluid.io/v1alpha1", Kind: "DataLoad"},
			}
			operation := testValueFileOperation{
				operationType: dataoperation.OperationType("Unknown"),
				object:        object,
			}
			engine := &CacheEngine{name: "runtime"}

			_, err := engine.GetDataOperationValueFile(cruntime.ReconcileRequestContext{Context: context.Background()}, operation)

			Expect(fluiderrors.IsNotSupported(err)).To(BeTrue())
			Expect(fluiderrors.ReasonForError(err)).To(Equal(fluiderrors.StatusReasonNotSupported))
			Expect(err.Error()).To(Equal("DataLoad is not supported by CacheRuntime[runtime]"))
		})
	})

	Describe("findDataOperationSpec", func() {
		It("returns the matching operation spec", func() {
			specs := []datav1alpha1.DataOperationSpec{
				{Name: string(dataoperation.DataLoadType)},
				{Name: string(dataoperation.DataProcessType), Command: []string{"process"}},
			}

			spec := findDataOperationSpec(specs, dataoperation.DataProcessType)

			Expect(spec).NotTo(BeNil())
			Expect(spec.Name).To(Equal(string(dataoperation.DataProcessType)))
			Expect(spec.Command).To(Equal([]string{"process"}))
		})

		It("returns nil when no operation spec matches", func() {
			spec := findDataOperationSpec(
				[]datav1alpha1.DataOperationSpec{{Name: string(dataoperation.DataLoadType)}},
				dataoperation.DataBackupType,
			)

			Expect(spec).To(BeNil())
		})
	})
})

type testValueFileOperation struct {
	dataoperation.OperationInterface
	operationType dataoperation.OperationType
	object        client.Object
}

func (o testValueFileOperation) GetOperationType() dataoperation.OperationType {
	return o.operationType
}

func (o testValueFileOperation) GetOperationObject() client.Object {
	return o.object
}
