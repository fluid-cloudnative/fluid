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

package dataoperation

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("OperationInterface contract", func() {
	It("should expose the operation label constant", func() {
		Expect(OperationLabel).To(Equal("fluid.io/operation"))
	})

	It("should expose the dataload operation type constant", func() {
		Expect(DataLoadType).To(Equal(OperationType("DataLoad")))
	})

	It("should expose the databackup operation type constant", func() {
		Expect(DataBackupType).To(Equal(OperationType("DataBackup")))
	})

	It("should expose the datamigrate operation type constant", func() {
		Expect(DataMigrateType).To(Equal(OperationType("DataMigrate")))
	})

	It("should expose the dataprocess operation type constant", func() {
		Expect(DataProcessType).To(Equal(OperationType("DataProcess")))
	})

	It("should build a dataload operation with the provided ttl", func() {
		ttl := int32(300)

		operation := BuildMockDataloadOperationReconcilerInterface(DataLoadType, &ttl)

		Expect(operation).NotTo(BeNil())
		Expect(operation.GetOperationType()).To(Equal(DataLoadType))
		existingTTL, err := operation.GetTTL()
		Expect(err).NotTo(HaveOccurred())
		Expect(existingTTL).To(Equal(&ttl))
		Expect(operation.GetParallelTaskNumber()).To(Equal(int32(1)))
	})

	It("should report a type mismatch when ttl is requested for the wrong operation type", func() {
		operation := BuildMockDataloadOperationReconcilerInterface(DataBackupType, nil)

		ttl, err := operation.GetTTL()

		Expect(err).To(MatchError("the dataoperation type is DataBackup, not DataloadType"))
		Expect(ttl).To(BeNil())
	})
})
