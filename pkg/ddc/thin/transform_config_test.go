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

package thin

import (
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("ThinEngine extractVolumeInfo", func() {
	var engine ThinEngine

	BeforeEach(func() {
		pvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pvc",
				Namespace: "fluid",
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				VolumeName: "test-pv",
			},
			Status: corev1.PersistentVolumeClaimStatus{
				Phase: corev1.ClaimBound,
			},
		}

		pv := &corev1.PersistentVolume{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-pv",
			},
			Spec: corev1.PersistentVolumeSpec{
				MountOptions: []string{"rw", "noexec"},
				PersistentVolumeSource: corev1.PersistentVolumeSource{
					CSI: &corev1.CSIPersistentVolumeSource{
						NodePublishSecretRef: &corev1.SecretReference{
							Name:      "my-secret",
							Namespace: "node-publish-secrets",
						},
						VolumeHandle: "test-pv",
						VolumeAttributes: map[string]string{
							"test-attr":  "true",
							"test-attr2": "foobar",
						},
					},
				},
			},
		}

		client := fake.NewFakeClientWithScheme(testScheme, pvc, pv)

		engine = ThinEngine{
			name:      "thin-test",
			namespace: "fluid",
			Client:    client,
			Log:       fake.NullLogger(),
		}
	})

	It("should extract volume info correctly", func() {
		wantCsiInfo := &corev1.CSIPersistentVolumeSource{
			NodePublishSecretRef: &corev1.SecretReference{
				Name:      "my-secret",
				Namespace: "node-publish-secrets",
			},
			VolumeHandle: "test-pv",
			VolumeAttributes: map[string]string{
				"test-attr":  "true",
				"test-attr2": "foobar",
			},
		}
		wantMountOptions := []string{"rw", "noexec"}

		gotCsiInfo, gotMountOptions, err := engine.extractVolumeInfo("test-pvc")
		Expect(err).NotTo(HaveOccurred())
		Expect(gotCsiInfo).To(Equal(wantCsiInfo))
		Expect(gotMountOptions).To(Equal(wantMountOptions))
	})
})

var _ = Describe("ThinEngine extractVolumeMountOptions", func() {
	var engine ThinEngine

	BeforeEach(func() {
		engine = ThinEngine{}
	})

	DescribeTable("extracting mount options from PV",
		func(pv *corev1.PersistentVolume, wantMountOptions []string, wantErr bool) {
			gotMountOptions, err := engine.extractVolumeMountOptions(pv)
			if wantErr {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(gotMountOptions).To(Equal(wantMountOptions))
			}
		},
		Entry("mount options in annotation",
			&corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						corev1.MountOptionAnnotation: "rw,noexec,testOpts",
					},
				},
			},
			[]string{"rw", "noexec", "testOpts"},
			false,
		),
		Entry("mount options in property",
			&corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
				Spec: corev1.PersistentVolumeSpec{
					MountOptions: []string{"ro", "noexec"},
				},
			},
			[]string{"ro", "noexec"},
			false,
		),
	)
})
