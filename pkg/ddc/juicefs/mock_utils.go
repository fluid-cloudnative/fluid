package juicefs

import (
	"fmt"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/tools/record"
)

type mockedObjects struct {
	WorkerSts  *appsv1.StatefulSet
	FuseDs     *appsv1.DaemonSet
	ConfigMaps map[string]*corev1.ConfigMap
	Services   map[string]*corev1.Service

	PersistentVolume      *corev1.PersistentVolume
	PersistentVolumeClaim *corev1.PersistentVolumeClaim
}

const mockedWorkerScript string = `
	#!/bin/bash
	set -e
	...
	exec /sbin/mount.juicefs juicefs-test /runtime-mnt/juicefs/%s/%s/juicefs-fuse -o cache-size=2048,free-space-ratio=0.1,cache-dir=%s,foreground,no-update,cache-group=%s-%s
`

const mockedFuseScript string = `
	#!/bin/bash
	set -e
	...
	exec /sbin/mount.juicefs juicefs-test /runtime-mnt/juicefs/%s/%s/juicefs-fuse -o cache-size=2048,free-space-ratio=0.1,cache-dir=%s,foreground,no-update,cache-group=%s-%s,no-sharing
`

func mockFluidObjectsForTests(namespacedName types.NamespacedName) (*datav1alpha1.Dataset, *datav1alpha1.JuiceFSRuntime) {
	datasetUid := uuid.NewUUID()
	dataset := &datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespacedName.Name,
			Namespace: namespacedName.Namespace,
			UID:       datasetUid,
			Labels: map[string]string{
				common.LabelAnnotationDatasetId: utils.GetDatasetId(namespacedName.Namespace, namespacedName.Name, string(datasetUid)),
			},
		},
		Spec: datav1alpha1.DatasetSpec{
			PlacementMode: datav1alpha1.DefaultMode,
			Mounts: []datav1alpha1.Mount{
				{
					MountPoint: "juicefs:///",
					Path:       "/",
					Name:       "juicefs-test",
					Options: map[string]string{
						"bucket": "https://mybucket.oss-cn-beijing-internal.aliyuncs.com",
					},
					EncryptOptions: []datav1alpha1.EncryptOption{
						{
							Name: "token",
							ValueFrom: datav1alpha1.EncryptOptionSource{
								SecretKeyRef: datav1alpha1.SecretKeySelector{
									Name: "jfs-secret",
									Key:  "token",
								},
							},
						},
						{
							Name: "access-key",
							ValueFrom: datav1alpha1.EncryptOptionSource{
								SecretKeyRef: datav1alpha1.SecretKeySelector{
									Name: "jfs-secret",
									Key:  "access-key",
								},
							},
						},
						{
							Name: "secret-key",
							ValueFrom: datav1alpha1.EncryptOptionSource{
								SecretKeyRef: datav1alpha1.SecretKeySelector{
									Name: "jfs-secret",
									Key:  "secret-key",
								},
							},
						},
					},
				},
			},
		},
	}

	tieredStoreQuota := resource.MustParse("2Gi")
	runtimeUid := uuid.NewUUID()
	runtime := &datav1alpha1.JuiceFSRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespacedName.Name,
			Namespace: namespacedName.Namespace,
			UID:       runtimeUid,
			Labels: map[string]string{
				common.LabelAnnotationDatasetId: utils.GetDatasetId(namespacedName.Namespace, dataset.Name, string(dataset.UID)),
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: datav1alpha1.GroupVersion.String(),
					Kind:       datav1alpha1.Datasetkind,
					Name:       namespacedName.Name,
					UID:        dataset.UID,
				},
			},
		},
		Spec: datav1alpha1.JuiceFSRuntimeSpec{
			Replicas: 2,
			TieredStore: datav1alpha1.TieredStore{
				Levels: []datav1alpha1.Level{
					{
						MediumType: "MEM",
						VolumeType: common.VolumeTypeEmptyDir,
						Path:       "/mnt/cache",
						Quota:      &tieredStoreQuota,
						Low:        "0.1",
					},
				},
			},
		},
	}

	return dataset, runtime
}

func mockJuiceFSEngineForTests(dataset *datav1alpha1.Dataset, runtime *datav1alpha1.JuiceFSRuntime) *JuiceFSEngine {
	if dataset.Namespace != runtime.Namespace ||
		dataset.Name != runtime.Name ||
		len(runtime.OwnerReferences) == 0 ||
		runtime.OwnerReferences[0].UID != dataset.UID {

		panic("engine mocking is only valid on a valid pair of dataset and runtime")
	}

	// build a very basic runtime info here. Test cases can override it with a new runtimeInfo.
	runtimeInfo, _ := base.BuildRuntimeInfo(dataset.Name, dataset.Namespace, common.JuiceFSRuntime)

	engine := &JuiceFSEngine{
		runtime:     runtime,
		name:        runtime.Name,
		namespace:   runtime.Namespace,
		runtimeType: common.JuiceFSRuntime,
		engineImpl:  common.JuiceFSEngineImpl,
		Log:         fake.NullLogger(),
		Recorder:    record.NewFakeRecorder(1),
		Client:      nil, // leave empty at this moment, test cases should handle this if Client is needed
		Helper:      nil, // leave empty at this moment, test cases should handle this if Helper is needed
		runtimeInfo: runtimeInfo,
		UnitTest:    true,
	}

	return engine
}

func mockJuiceFSObjectsForTests(dataset *datav1alpha1.Dataset, runtime *datav1alpha1.JuiceFSRuntime, engine *JuiceFSEngine) mockedObjects {
	if dataset.Namespace != runtime.Namespace ||
		dataset.Name != runtime.Name ||
		len(runtime.OwnerReferences) == 0 ||
		runtime.OwnerReferences[0].UID != dataset.UID {

		panic("juicefs objects mocking is only valid on a valid pair of dataset and runtime")
	}

	commonLabels := map[string]string{
		"app":                          "juicefs",
		"chart":                        "juicefs-0.9.13",
		"release":                      runtime.Name,
		"heritage":                     "Helm",
		"app.kubernetes.io/name":       "juicefs",
		"app.kubernetes.io/instance":   runtime.Name,
		"app.kubernetes.io/managed-by": "Helm",
		"fluid.io/managed-by":          "fluid",
		"fluid.io/dataset-id":          utils.GetDatasetId(dataset.Namespace, dataset.Name, string(dataset.UID)),
	}

	configMaps := map[string]*corev1.ConfigMap{}
	// mock worker sts
	workerLabels := map[string]string{
		"role":             "juicefs-worker",
		"name":             engine.name + "-worker",
		"fluid.io/dataset": utils.GetDatasetId(dataset.Namespace, dataset.Name, string(dataset.UID)),
	}
	for k, v := range commonLabels {
		workerLabels[k] = v
	}
	workerSts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      engine.name + "-worker",
			Namespace: runtime.Namespace,
			Labels:    workerLabels,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: datav1alpha1.GroupVersion.String(),
					Kind:       datav1alpha1.JuiceFSRuntimeKind,
					Name:       runtime.Name,
					UID:        runtime.UID,
				},
			},
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &runtime.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":  workerLabels["app"],
					"role": workerLabels["role"],
					"name": workerLabels["name"],
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: workerLabels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: JuiceFSWorkerContainerName,
						},
					},
				},
			},
		},
	}

	workerScriptCm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-worker-script", runtime.Name),
			Namespace: runtime.Namespace,
			Labels:    workerLabels,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: datav1alpha1.GroupVersion.String(),
					Kind:       datav1alpha1.JuiceFSRuntimeKind,
					Name:       runtime.Name,
					UID:        runtime.UID,
				},
			},
		},
		Data: map[string]string{
			"script.sh": fmt.Sprintf(mockedWorkerScript, runtime.Namespace, runtime.Name, runtime.Spec.TieredStore.Levels[0].Path, runtime.Namespace, runtime.Name),
		},
	}
	configMaps[workerScriptCm.Name] = workerScriptCm

	// mock fuse ds
	fuseLabels := map[string]string{
		"role": "juicefs-fuse",
		"name": engine.name + "-fuse",
	}
	for k, v := range commonLabels {
		fuseLabels[k] = v
	}

	fuseDs := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      engine.name + "-fuse",
			Namespace: runtime.Namespace,
			Labels:    fuseLabels,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: datav1alpha1.GroupVersion.String(),
					Kind:       datav1alpha1.JuiceFSRuntimeKind,
					Name:       runtime.Name,
					UID:        runtime.UID,
				},
			},
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":      fuseLabels["app"],
					"chart":    fuseLabels["chart"],
					"release":  fuseLabels["release"],
					"heritage": fuseLabels["heritage"],
					"role":     fuseLabels["role"],
				},
			},
			UpdateStrategy: appsv1.DaemonSetUpdateStrategy{
				Type: appsv1.OnDeleteDaemonSetStrategyType,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: fuseLabels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: JuiceFSFuseContainerName,
						},
					},
				},
			},
		},
	}

	fuseScriptCm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-fuse-script", runtime.Name),
			Namespace: runtime.Namespace,
			Labels:    fuseLabels,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: datav1alpha1.GroupVersion.String(),
					Kind:       datav1alpha1.JuiceFSRuntimeKind,
					Name:       runtime.Name,
					UID:        runtime.UID,
				},
			},
		},
		Data: map[string]string{
			"script.sh": fmt.Sprintf(mockedFuseScript, runtime.Namespace, runtime.Name, runtime.Spec.TieredStore.Levels[0].Path, runtime.Namespace, runtime.Name),
		},
	}
	configMaps[fuseScriptCm.Name] = fuseScriptCm

	// mock persistent volume
	accessModes := []corev1.PersistentVolumeAccessMode{corev1.ReadOnlyMany}
	if len(dataset.Spec.AccessModes) != 0 {
		accessModes = dataset.Spec.AccessModes
	}
	mountPath := "/runtime-mnt/juicefs/" + runtime.Namespace + "/" + runtime.Name + "/juicefs-fuse"
	pv := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", runtime.Namespace, runtime.Name),
			Namespace: runtime.Namespace,
			Labels: map[string]string{
				common.LabelAnnotationDatasetId: utils.GetDatasetId(runtime.Namespace, runtime.Name, string(dataset.UID)),
				utils.GetCommonLabelName(runtime.Namespace, runtime.Name, string(dataset.UID)): "true",
			},
			Annotations: common.GetExpectedFluidAnnotations(),
		},
		Spec: corev1.PersistentVolumeSpec{
			ClaimRef: &corev1.ObjectReference{
				Namespace: runtime.Namespace,
				Name:      runtime.Name,
			},
			AccessModes: accessModes,
			Capacity: corev1.ResourceList{
				corev1.ResourceName(corev1.ResourceStorage): resource.MustParse(utils.DefaultStorageCapacity),
			},
			StorageClassName: common.FluidStorageClass,
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				CSI: &corev1.CSIPersistentVolumeSource{
					Driver:       common.CSIDriver,
					VolumeHandle: fmt.Sprintf("%s-%s", runtime.Namespace, runtime.Name),
					VolumeAttributes: map[string]string{
						common.VolumeAttrFluidPath: mountPath,
						common.VolumeAttrMountType: common.JuiceFSMountType,
						common.VolumeAttrNamespace: runtime.Namespace,
						common.VolumeAttrName:      runtime.Name,
					},
				},
			},
		},
	}

	// mock persistent volume claim
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      runtime.Name,
			Namespace: runtime.Namespace,
			Labels: map[string]string{
				utils.GetCommonLabelName(runtime.Namespace, runtime.Name, string(dataset.UID)): "true",
				common.LabelAnnotationDatasetId: utils.GetDatasetId(runtime.Namespace, runtime.Name, string(dataset.UID)),
			},
			Annotations: common.GetExpectedFluidAnnotations(),
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			VolumeName:       fmt.Sprintf("%s-%s", runtime.Namespace, runtime.Name),
			StorageClassName: &common.FluidStorageClass,
			AccessModes:      accessModes,
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(utils.DefaultStorageCapacity),
				},
			},
		},
	}

	return mockedObjects{
		WorkerSts:  workerSts,
		FuseDs:     fuseDs,
		ConfigMaps: configMaps,

		PersistentVolume:      pv,
		PersistentVolumeClaim: pvc,
	}
}

func mockJuiceFSValue(dataset *datav1alpha1.Dataset, juicefsruntime *datav1alpha1.JuiceFSRuntime) *JuiceFS {
	value := &JuiceFS{
		Fuse: Fuse{
			HostNetwork: true,
			NodeSelector: map[string]string{
				fmt.Sprintf("fluid.io/f-%s-%s", juicefsruntime.Namespace, juicefsruntime.Name): "true",
			},
			VolumeMounts: []corev1.VolumeMount{
				{MountPath: juicefsruntime.Spec.TieredStore.Levels[0].Path, Name: "cache-dir-0"},
			},
			Volumes: []corev1.Volume{
				{Name: "cache-dir-0", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
			},
			Command: fmt.Sprintf(mockedFuseScript, juicefsruntime.Namespace, juicefsruntime.Name, juicefsruntime.Spec.TieredStore.Levels[0].Path, juicefsruntime.Namespace, juicefsruntime.Name),
		},
		Worker: Worker{
			HostNetwork: true,
			VolumeMounts: []corev1.VolumeMount{
				{MountPath: juicefsruntime.Spec.TieredStore.Levels[0].Path, Name: "cache-dir-0"},
			},
			Volumes: []corev1.Volume{
				{Name: "cache-dir-0", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
			},
			Command: fmt.Sprintf(mockedWorkerScript, juicefsruntime.Namespace, juicefsruntime.Name, juicefsruntime.Spec.TieredStore.Levels[0].Path, juicefsruntime.Namespace, juicefsruntime.Name),
		},
	}

	if juicefsruntime.Spec.TieredStore.Levels[0].MediumType == common.Memory {
		value.Fuse.Resources.Requests = common.ResourceList{}
		value.Fuse.Resources.Requests[corev1.ResourceMemory] = juicefsruntime.Spec.TieredStore.Levels[0].Quota.String()
		value.Worker.Resources.Requests = common.ResourceList{}
		value.Worker.Resources.Requests[corev1.ResourceMemory] = juicefsruntime.Spec.TieredStore.Levels[0].Quota.String()
	}

	return value
}
