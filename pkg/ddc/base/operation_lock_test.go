/*
Copyright 2023 The Fluid Authors.

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

package base

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("getDataOperationKey", func() {
	It("should return the object name", func() {
		obj := &datav1alpha1.DataBackup{
			ObjectMeta: metav1.ObjectMeta{Name: "my-backup", Namespace: "default"},
		}
		Expect(getDataOperationKey(obj)).To(Equal("my-backup"))
	})

	It("should return empty string for object with no name", func() {
		obj := &datav1alpha1.DataBackup{
			ObjectMeta: metav1.ObjectMeta{Name: "", Namespace: "default"},
		}
		Expect(getDataOperationKey(obj)).To(Equal(""))
	})

	It("should return empty string for nil object", func() {
		var obj *datav1alpha1.DataBackup = nil
		Expect(getDataOperationKey(obj)).To(Equal(""))
	})
})
