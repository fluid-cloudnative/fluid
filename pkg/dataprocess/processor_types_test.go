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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("ValidateDatasetMountPath", func() {

	Describe("ScriptProcessorImpl", func() {

		Context("when validating dataset mount path", func() {

			It("should fail when volume mount path conflicts", func() {
				p := &ScriptProcessorImpl{
					ScriptProcessor: &datav1alpha1.ScriptProcessor{
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "myvol",
								MountPath: "/fluid-data",
							},
						},
					},
				}

				pass, conflictVol, conflictContainer :=
					p.ValidateDatasetMountPath("/fluid-data")

				Expect(pass).To(BeFalse())
				Expect(conflictVol).To(Equal("myvol"))
				Expect(conflictContainer).
					To(Equal(DataProcessScriptProcessorContainerName))
			})

			It("should pass when no conflict exists", func() {
				p := &ScriptProcessorImpl{
					ScriptProcessor: &datav1alpha1.ScriptProcessor{
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "myvol",
								MountPath: "/mydata",
							},
						},
					},
				}

				pass, conflictVol, conflictContainer :=
					p.ValidateDatasetMountPath("/fluid-data")

				Expect(pass).To(BeTrue())
				Expect(conflictVol).To(BeEmpty())
				Expect(conflictContainer).To(BeEmpty())
			})
		})
	})

	Describe("JobProcessorImpl", func() {

		Context("when container volume mount conflicts", func() {
			It("should fail and return container name", func() {
				p := &JobProcessorImpl{
					JobProcessor: &datav1alpha1.JobProcessor{
						PodSpec: &corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "test-container",
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "myvol",
											MountPath: "/fluid-data",
										},
									},
								},
							},
						},
					},
				}

				pass, conflictVol, conflictContainer :=
					p.ValidateDatasetMountPath("/fluid-data")

				Expect(pass).To(BeFalse())
				Expect(conflictVol).To(Equal("myvol"))
				Expect(conflictContainer).To(Equal("test-container"))
			})
		})

		Context("when init container volume mount conflicts", func() {
			It("should fail and return init container name", func() {
				p := &JobProcessorImpl{
					JobProcessor: &datav1alpha1.JobProcessor{
						PodSpec: &corev1.PodSpec{
							InitContainers: []corev1.Container{
								{
									Name: "test-init-container",
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "myvol",
											MountPath: "/fluid-data",
										},
									},
								},
							},
						},
					},
				}

				pass, conflictVol, conflictContainer :=
					p.ValidateDatasetMountPath("/fluid-data")

				Expect(pass).To(BeFalse())
				Expect(conflictVol).To(Equal("myvol"))
				Expect(conflictContainer).To(Equal("test-init-container"))
			})
		})

		Context("when no container has conflicting mount path", func() {
			It("should pass validation", func() {
				p := &JobProcessorImpl{
					JobProcessor: &datav1alpha1.JobProcessor{
						PodSpec: &corev1.PodSpec{
							InitContainers: []corev1.Container{
								{
									Name: "test-init-container",
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "myvol",
											MountPath: "/mydata",
										},
									},
								},
							},
							Containers: []corev1.Container{
								{
									Name: "test-container",
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "myvol",
											MountPath: "/mydata",
										},
									},
								},
							},
						},
					},
				}

				pass, conflictVol, conflictContainer :=
					p.ValidateDatasetMountPath("/fluid-data")

				Expect(pass).To(BeTrue())
				Expect(conflictVol).To(BeEmpty())
				Expect(conflictContainer).To(BeEmpty())
			})
		})
	})
})

var _ = Describe("GetProcessorImpl", func() {

	It("should return JobProcessorImpl when JobProcessor is set", func() {
		dp := &datav1alpha1.DataProcess{
			Spec: datav1alpha1.DataProcessSpec{
				Processor: datav1alpha1.Processor{
					Job: &datav1alpha1.JobProcessor{},
				},
			},
		}

		processor := GetProcessorImpl(dp)
		Expect(processor).To(BeAssignableToTypeOf(&JobProcessorImpl{}))
	})

	It("should return ScriptProcessorImpl when ScriptProcessor is set", func() {
		dp := &datav1alpha1.DataProcess{
			Spec: datav1alpha1.DataProcessSpec{
				Processor: datav1alpha1.Processor{
					Script: &datav1alpha1.ScriptProcessor{},
				},
			},
		}

		processor := GetProcessorImpl(dp)
		Expect(processor).To(BeAssignableToTypeOf(&ScriptProcessorImpl{}))
	})

	It("should return nil when no processor is set", func() {
		dp := &datav1alpha1.DataProcess{
			Spec: datav1alpha1.DataProcessSpec{},
		}

		processor := GetProcessorImpl(dp)
		Expect(processor).To(BeNil())
	})
})

var _ = Describe("JobProcessorImpl TransformDataProcessValues", func() {

	It("should append volumes and volume mounts to containers and init containers", func() {
		p := &JobProcessorImpl{
			JobProcessor: &datav1alpha1.JobProcessor{
				PodSpec: &corev1.PodSpec{
					Containers:     []corev1.Container{{Name: "c1"}},
					InitContainers: []corev1.Container{{Name: "ic1"}},
				},
			},
		}

		value := &DataProcessValue{}
		volumes := []corev1.Volume{{Name: "dataset-vol"}}
		mounts := []corev1.VolumeMount{{Name: "dataset-vol", MountPath: "/data"}}

		p.TransformDataProcessValues(value, volumes, mounts)

		Expect(value.DataProcessInfo.JobProcessor).NotTo(BeNil())
		Expect(value.DataProcessInfo.JobProcessor.PodSpec.Volumes).To(ContainElement(volumes[0]))
		Expect(value.DataProcessInfo.JobProcessor.PodSpec.Containers[0].VolumeMounts).
			To(ContainElement(mounts[0]))
		Expect(value.DataProcessInfo.JobProcessor.PodSpec.InitContainers[0].VolumeMounts).
			To(ContainElement(mounts[0]))
	})

	It("should not panic when no containers exist", func() {
		p := &JobProcessorImpl{
			JobProcessor: &datav1alpha1.JobProcessor{
				PodSpec: &corev1.PodSpec{},
			},
		}

		value := &DataProcessValue{}
		Expect(func() {
			p.TransformDataProcessValues(value, nil, nil)
		}).NotTo(Panic())

		Expect(value.DataProcessInfo.JobProcessor).NotTo(BeNil())
	})
})
