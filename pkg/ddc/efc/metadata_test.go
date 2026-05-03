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

package efc

import (
	"context"
	"errors"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("EFCEngine metadata", func() {
	BeforeEach(func() {
		shouldCheckUFS = func(e *EFCEngine) (bool, error) {
			return e.ShouldCheckUFS()
		}
		totalStorageBytes = func(e *EFCEngine) (int64, error) {
			return e.TotalStorageBytes()
		}
		totalFileNums = func(e *EFCEngine) (int64, error) {
			return e.TotalFileNums()
		}
	})

	Describe("syncMetadataInternal", func() {
		It("updates dataset metadata using the current UFS totals", func() {
			dataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "spark",
					Namespace: "fluid",
				},
			}

			engine := &EFCEngine{
				name:      "spark",
				namespace: "fluid",
				Client:    fake.NewFakeClientWithScheme(testScheme, dataset.DeepCopy()),
				Log:       fake.NullLogger(),
			}

			Expect(engine.syncMetadataInternal()).To(Succeed())

			updated := &datav1alpha1.Dataset{}
			Expect(engine.Client.Get(context.TODO(), types.NamespacedName{Name: "spark", Namespace: "fluid"}, updated)).To(Succeed())
			Expect(updated.Status.UfsTotal).To(Equal("0.00B"))
			Expect(updated.Status.FileNum).To(Equal("0"))
		})

		It("returns an error when querying total storage fails", func() {
			expectedErr := errors.New("storage failure")
			totalStorageBytes = func(*EFCEngine) (int64, error) {
				return 0, expectedErr
			}

			engine := &EFCEngine{
				name:      "spark",
				namespace: "fluid",
				Client:    fake.NewFakeClientWithScheme(testScheme),
				Log:       fake.NullLogger(),
			}

			Expect(engine.syncMetadataInternal()).To(MatchError(expectedErr))
		})

		It("returns an error when querying total file count fails", func() {
			expectedErr := errors.New("file count failure")
			totalFileNums = func(*EFCEngine) (int64, error) {
				return 0, expectedErr
			}

			engine := &EFCEngine{
				name:      "spark",
				namespace: "fluid",
				Client:    fake.NewFakeClientWithScheme(testScheme),
				Log:       fake.NullLogger(),
			}

			Expect(engine.syncMetadataInternal()).To(MatchError(expectedErr))
		})

		It("returns an error when the dataset cannot be loaded", func() {
			engine := &EFCEngine{
				name:      "spark",
				namespace: "fluid",
				Client:    fake.NewFakeClientWithScheme(testScheme),
				Log:       fake.NullLogger(),
			}

			err := engine.syncMetadataInternal()

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("datasets.data.fluid.io \"spark\" not found"))
		})
	})

	Describe("SyncMetadata", func() {
		It("syncs metadata when UFS checks are enabled", func() {
			dataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "spark",
					Namespace: "fluid",
				},
			}

			shouldCheckUFS = func(*EFCEngine) (bool, error) {
				return true, nil
			}
			totalStorageBytes = func(*EFCEngine) (int64, error) {
				return 1024, nil
			}
			totalFileNums = func(*EFCEngine) (int64, error) {
				return 7, nil
			}

			engine := &EFCEngine{
				name:      "spark",
				namespace: "fluid",
				Client:    fake.NewFakeClientWithScheme(testScheme, dataset.DeepCopy()),
				Log:       fake.NullLogger(),
			}

			Expect(engine.SyncMetadata()).To(Succeed())

			updated := &datav1alpha1.Dataset{}
			Expect(engine.Client.Get(context.TODO(), types.NamespacedName{Name: "spark", Namespace: "fluid"}, updated)).To(Succeed())
			Expect(updated.Status.UfsTotal).To(Equal("1.00KiB"))
			Expect(updated.Status.FileNum).To(Equal("7"))
		})

		It("skips syncing when the engine does not need UFS metadata checks", func() {
			dataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "spark",
					Namespace: "fluid",
				},
				Status: datav1alpha1.DatasetStatus{
					UfsTotal: "existing-total",
					FileNum:  "existing-files",
				},
			}

			engine := &EFCEngine{
				name:      "spark",
				namespace: "fluid",
				Client:    fake.NewFakeClientWithScheme(testScheme, dataset.DeepCopy()),
				Log:       fake.NullLogger(),
			}

			Expect(engine.SyncMetadata()).To(Succeed())

			unchanged := &datav1alpha1.Dataset{}
			Expect(engine.Client.Get(context.TODO(), types.NamespacedName{Name: "spark", Namespace: "fluid"}, unchanged)).To(Succeed())
			Expect(unchanged.Status.UfsTotal).To(Equal("existing-total"))
			Expect(unchanged.Status.FileNum).To(Equal("existing-files"))
		})
	})
})
