package mutator

import (
	"fmt"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMutator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mutator Suite")
}

func test_buildPodToMutate(pvcNames []string) *corev1.Pod {
	volumes := []corev1.Volume{}
	volumeMounts := []corev1.VolumeMount{}
	for i, pvcName := range pvcNames {
		volumeName := fmt.Sprintf("data-vol-%d", i)
		volumes = append(volumes, corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: pvcName,
				},
			},
		})

		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      volumeName,
			MountPath: fmt.Sprintf("/data%d", i),
		})
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:         "test-ctr",
					Image:        "test-image",
					VolumeMounts: volumeMounts,
				},
			},
			Volumes: volumes,
		},
	}

	return pod
}

func test_buildFluidResources(datasetName string, datasetNamespace string) (*datav1alpha1.Dataset, *datav1alpha1.ThinRuntime, *appsv1.DaemonSet, *corev1.PersistentVolume) {
	dataset := &datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      datasetName,
			Namespace: datasetNamespace,
		},
		Spec: datav1alpha1.DatasetSpec{},
		Status: datav1alpha1.DatasetStatus{
			Runtimes: []datav1alpha1.Runtime{
				{
					Type: common.ThinRuntime,
				},
			},
		},
	}

	thinRuntime := &datav1alpha1.ThinRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      datasetName,
			Namespace: datasetNamespace,
		},
		Spec:   datav1alpha1.ThinRuntimeSpec{},
		Status: datav1alpha1.RuntimeStatus{},
	}

	mountPropagationBidirectional := corev1.MountPropagationBidirectional
	daemonSet := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-fuse", datasetName),
			Namespace: datasetNamespace,
		},
		Spec: appsv1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "fuse-ctr",
							Image: "myimage:fuse",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:             "thin-fuse-mount",
									MountPath:        fmt.Sprintf("/runtime-mnt/thin/%s/%s/", datasetNamespace, datasetName),
									MountPropagation: &mountPropagationBidirectional,
								},
							},
							SecurityContext: &corev1.SecurityContext{
								Privileged: ptr.To(true),
								Capabilities: &corev1.Capabilities{
									Add: []corev1.Capability{
										"SYS_ADMIN",
									},
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "thin-fuse-mount",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: fmt.Sprintf("/runtime-mnt/thin/%s/%s/", datasetNamespace, datasetName),
								},
							},
						},
					},
				},
			},
		},
	}

	pv := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s", datasetNamespace, datasetName),
		},
		Spec: corev1.PersistentVolumeSpec{
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				CSI: &corev1.CSIPersistentVolumeSource{
					VolumeAttributes: map[string]string{
						common.VolumeAttrFluidPath: fmt.Sprintf("/runtime-mnt/thin/%s/%s/thin-fuse", datasetNamespace, datasetName),
						common.VolumeAttrMountType: "myfuse-fstype",
					},
				},
			},
		},
	}

	return dataset, thinRuntime, daemonSet, pv
}
