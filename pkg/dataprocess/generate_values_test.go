/*
Copyright 2023 The Fluid Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0
*/

package dataprocess

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/transformer"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
)

var _ = Describe("GenDataProcessValue", func() {

	var (
		dataset *datav1alpha1.Dataset
	)

	BeforeEach(func() {
		dataset = &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "demo-dataset",
				Namespace: "default",
			},
		}
	})

	Describe("transformCommonPart", func() {

		It("should set name, labels, annotations, owner and serviceAccount", func() {
			dp := &datav1alpha1.DataProcess{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "dp",
					Namespace:   "default",
					Annotations: map[string]string{"a": "b"},
				},
				Spec: datav1alpha1.DataProcessSpec{
					Processor: datav1alpha1.Processor{
						ServiceAccountName: "sa-test",
						PodMetadata: datav1alpha1.PodMetadata{
							Labels: map[string]string{"l": "v"},
							Annotations: map[string]string{
								"pod": "anno",
							},
						},
					},
				},
			}

			value := &DataProcessValue{
				DataProcessInfo: DataProcessInfo{},
			}

			transformCommonPart(value, dp)

			Expect(value.Name).To(Equal("dp"))
			Expect(value.DataProcessInfo.Labels).To(Equal(map[string]string{"l": "v"}))
			Expect(value.DataProcessInfo.ServiceAccountName).To(Equal("sa-test"))
			Expect(value.Owner).To(Equal(transformer.GenerateOwnerReferenceFromObject(dp)))
			Expect(value.DataProcessInfo.Annotations).To(HaveKey("pod"))
		})
	})

	Describe("GenDataProcessValueFile", func() {

		It("should generate value file for ScriptProcessor", func() {
			dataProcess := &datav1alpha1.DataProcess{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "dp",
					Namespace: "default",
				},
				Spec: datav1alpha1.DataProcessSpec{
					Dataset: datav1alpha1.TargetDatasetWithMountPath{
						TargetDataset: datav1alpha1.TargetDataset{
							Name: dataset.Name,
						},
					},
					Processor: datav1alpha1.Processor{
						Script: &datav1alpha1.ScriptProcessor{
							VersionSpec: datav1alpha1.VersionSpec{
								Image: "busybox",
							},
						},
					},
				},
			}

			// fake client is enough; affinity injection allows nil
			file, err := GenDataProcessValueFile(nil, dataset, dataProcess)

			Expect(err).NotTo(HaveOccurred())
			Expect(file).NotTo(BeEmpty())

			_, statErr := os.Stat(file)
			Expect(statErr).NotTo(HaveOccurred())
		})
	})

	It("should generate value file for JobProcessor", func() {
		dataProcess := &datav1alpha1.DataProcess{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dp-job",
				Namespace: "default",
			},
			Spec: datav1alpha1.DataProcessSpec{
				Dataset: datav1alpha1.TargetDatasetWithMountPath{
					TargetDataset: datav1alpha1.TargetDataset{
						Name: dataset.Name,
					},
				},
				Processor: datav1alpha1.Processor{
					Job: &datav1alpha1.JobProcessor{
						PodSpec: &corev1.PodSpec{
							Containers: []corev1.Container{
								{Name: "c"},
							},
						},
					},
				},
			},
		}

		file, err := GenDataProcessValueFile(nil, dataset, dataProcess)

		Expect(err).NotTo(HaveOccurred())
		Expect(file).NotTo(BeEmpty())
	})

	It("should handle empty RunAfter", func() {
		dp := &datav1alpha1.DataProcess{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dp",
				Namespace: "default",
			},
			Spec: datav1alpha1.DataProcessSpec{
				Processor: datav1alpha1.Processor{
					Script: &datav1alpha1.ScriptProcessor{
						VersionSpec: datav1alpha1.VersionSpec{
							Image: "busybox",
						},
					},
				},
			},
		}

		Expect(func() {
			GenDataProcessValue(dataset, dp)
		}).NotTo(Panic())
	})

	Describe("ScriptProcessor", func() {

		It("should generate DataProcessValue with dataset volume and mount", func() {
			dataProcess := &datav1alpha1.DataProcess{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "demo-process",
					Namespace: "default",
				},
				Spec: datav1alpha1.DataProcessSpec{
					Dataset: datav1alpha1.TargetDatasetWithMountPath{
						TargetDataset: datav1alpha1.TargetDataset{
							Name:      dataset.Name,
							Namespace: dataset.Namespace,
						},
						MountPath: "/data",
					},
					Processor: datav1alpha1.Processor{
						Script: &datav1alpha1.ScriptProcessor{
							VersionSpec: datav1alpha1.VersionSpec{
								Image:           "test-image",
								ImageTag:        "latest",
								ImagePullPolicy: "IfNotPresent",
							},
							RestartPolicy: corev1.RestartPolicyNever,
							Command:       []string{"bash"},
							Source:        "sleep inf",
							Env: []corev1.EnvVar{
								{
									Name:  "TEST_ENV",
									Value: "foobar",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "mycode",
									MountPath: "/code",
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "mycode",
									VolumeSource: corev1.VolumeSource{
										PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
											ClaimName: "mypvc",
										},
									},
								},
							},
						},
					},
				},
			}

			want := &DataProcessValue{
				Name:  dataProcess.Name,
				Owner: transformer.GenerateOwnerReferenceFromObject(dataProcess),
				DataProcessInfo: DataProcessInfo{
					TargetDataset: dataset.Name,
					JobProcessor:  nil,
					ScriptProcessor: &ScriptProcessor{
						Image:           "test-image:latest",
						ImagePullPolicy: "IfNotPresent",
						RestartPolicy:   dataProcess.Spec.Processor.Script.RestartPolicy,
						Envs:            dataProcess.Spec.Processor.Script.Env,
						Command:         dataProcess.Spec.Processor.Script.Command,
						Source:          dataProcess.Spec.Processor.Script.Source,
						Volumes: append(
							dataProcess.Spec.Processor.Script.Volumes,
							corev1.Volume{
								Name: "fluid-dataset-vol",
								VolumeSource: corev1.VolumeSource{
									PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
										ClaimName: dataset.Name,
									},
								},
							},
						),
						VolumeMounts: append(
							dataProcess.Spec.Processor.Script.VolumeMounts,
							corev1.VolumeMount{
								Name:      "fluid-dataset-vol",
								MountPath: "/data",
							},
						),
					},
				},
			}

			got := GenDataProcessValue(dataset, dataProcess)
			Expect(got).To(Equal(want))
		})
	})

	Describe("JobProcessor", func() {

		It("should generate DataProcessValue with dataset volume injected", func() {
			dataProcess := &datav1alpha1.DataProcess{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "demo-process",
					Namespace: "default",
				},
				Spec: datav1alpha1.DataProcessSpec{
					Dataset: datav1alpha1.TargetDatasetWithMountPath{
						TargetDataset: datav1alpha1.TargetDataset{
							Name:      dataset.Name,
							Namespace: dataset.Namespace,
						},
						MountPath: "/data",
					},
					Processor: datav1alpha1.Processor{
						Job: &datav1alpha1.JobProcessor{
							PodSpec: &corev1.PodSpec{
								RestartPolicy: corev1.RestartPolicyOnFailure,
								Containers: []corev1.Container{
									{
										Image:           "test-image",
										ImagePullPolicy: "IfNotPresent",
									},
								},
							},
						},
					},
				},
			}

			modifiedPodSpec := &corev1.PodSpec{
				RestartPolicy: corev1.RestartPolicyOnFailure,
				Containers: []corev1.Container{
					{
						Image:           "test-image",
						ImagePullPolicy: "IfNotPresent",
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "fluid-dataset-vol",
								MountPath: "/data",
							},
						},
					},
				},
				Volumes: []corev1.Volume{
					{
						Name: "fluid-dataset-vol",
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: dataset.Name,
							},
						},
					},
				},
			}

			want := &DataProcessValue{
				Name:  dataProcess.Name,
				Owner: transformer.GenerateOwnerReferenceFromObject(dataProcess),
				DataProcessInfo: DataProcessInfo{
					TargetDataset:   dataset.Name,
					ScriptProcessor: nil,
					JobProcessor: &JobProcessor{
						PodSpec: modifiedPodSpec,
					},
				},
			}

			got := GenDataProcessValue(dataset, dataProcess)
			Expect(got).To(Equal(want))
		})
	})

	Describe("Without dataset mount path", func() {

		It("should not inject dataset volume for ScriptProcessor", func() {
			dataProcess := (&datav1alpha1.DataProcess{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "demo-process",
					Namespace: "default",
				},
				Spec: datav1alpha1.DataProcessSpec{
					Dataset: datav1alpha1.TargetDatasetWithMountPath{
						TargetDataset: datav1alpha1.TargetDataset{
							Name:      dataset.Name,
							Namespace: dataset.Namespace,
						},
					},
					Processor: datav1alpha1.Processor{
						Script: &datav1alpha1.ScriptProcessor{
							VersionSpec: datav1alpha1.VersionSpec{
								Image:           "test-image",
								ImageTag:        "latest",
								ImagePullPolicy: "IfNotPresent",
							},
							RestartPolicy: corev1.RestartPolicyNever,
						},
					},
				},
			})

			got := GenDataProcessValue(dataset, dataProcess)

			Expect(got.DataProcessInfo.ScriptProcessor.Volumes).To(BeEmpty())
			Expect(got.DataProcessInfo.ScriptProcessor.VolumeMounts).To(BeEmpty())
		})

		It("should not inject dataset volume for JobProcessor", func() {
			dataProcess := (&datav1alpha1.DataProcess{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "demo-process",
					Namespace: "default",
				},
				Spec: datav1alpha1.DataProcessSpec{
					Dataset: datav1alpha1.TargetDatasetWithMountPath{
						TargetDataset: datav1alpha1.TargetDataset{
							Name:      dataset.Name,
							Namespace: dataset.Namespace,
						},
					},
					Processor: datav1alpha1.Processor{
						Job: &datav1alpha1.JobProcessor{
							PodSpec: &corev1.PodSpec{
								RestartPolicy: corev1.RestartPolicyOnFailure,
								Containers: []corev1.Container{
									{
										Image:           "test-image",
										ImagePullPolicy: "IfNotPresent",
									},
								},
							},
						},
					},
				},
			})

			got := GenDataProcessValue(dataset, dataProcess)

			Expect(got.DataProcessInfo.JobProcessor.PodSpec.Volumes).To(BeNil())
			Expect(got.DataProcessInfo.JobProcessor.PodSpec.Containers[0].VolumeMounts).To(BeNil())
		})
	})
})
