package jindocache

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	// appsv1 "k8s.io/api/apps/v1"
	// corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/tools/record"
)

// type mockedObjects struct {
// 	MasterSts  *appsv1.StatefulSet
// 	WorkerSts  *appsv1.StatefulSet
// 	FuseDs     *appsv1.DaemonSet
// 	ConfigMaps map[string]*corev1.ConfigMap
// 	Services   map[string]*corev1.Service

// 	PersistentVolume      *corev1.PersistentVolume
// 	PersistentVolumeClaim *corev1.PersistentVolumeClaim
// }

func mockFluidObjectsForTests(namespacedName types.NamespacedName) (*datav1alpha1.Dataset, *datav1alpha1.JindoRuntime) {
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
					MountPoint: "oss://mybucket/mypath",
					Path:       "/",
					Name:       "oss-bucket",
					Options: map[string]string{
						"fs.oss.endpoint": "oss-cn-beijing-internal.aliyuncs.com",
					},
					EncryptOptions: []datav1alpha1.EncryptOption{
						{
							Name: "fs.oss.accessKeyId",
							ValueFrom: datav1alpha1.EncryptOptionSource{
								SecretKeyRef: datav1alpha1.SecretKeySelector{
									Name: "my-oss-secret",
									Key:  "myak",
								},
							},
						},
						{
							Name: "fs.oss.accessKeySecret",
							ValueFrom: datav1alpha1.EncryptOptionSource{
								SecretKeyRef: datav1alpha1.SecretKeySelector{
									Name: "my-oss-secret",
									Key:  "mysk",
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
	runtime := &datav1alpha1.JindoRuntime{
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
		Spec: datav1alpha1.JindoRuntimeSpec{
			Replicas: 2,
			TieredStore: datav1alpha1.TieredStore{
				Levels: []datav1alpha1.Level{
					{
						MediumType: "MEM",
						Path:       "/dev/shm",
						Quota:      &tieredStoreQuota,
						High:       "0.99",
						Low:        "0.95",
					},
				},
			},
		},
	}

	return dataset, runtime
}

func mockJindoCacheEngineForTests(dataset *datav1alpha1.Dataset, runtime *datav1alpha1.JindoRuntime) *JindoCacheEngine {
	if dataset.Namespace != runtime.Namespace ||
		dataset.Name != runtime.Name ||
		len(runtime.OwnerReferences) == 0 ||
		runtime.OwnerReferences[0].UID != dataset.UID {

		panic("engine mocking is only valid on a valid pair of dataset and runtime")
	}

	// build a very basic runtime info here. Test cases can override it with a new runtimeInfo.
	runtimeInfo, _ := base.BuildRuntimeInfo(dataset.Name, dataset.Namespace, common.JindoRuntime)

	engine := &JindoCacheEngine{
		runtime:     runtime,
		name:        runtime.Name,
		namespace:   runtime.Namespace,
		runtimeType: common.JindoRuntime,
		engineImpl:  common.JindoCacheEngineImpl,
		Log:         fake.NullLogger(),
		Recorder:    record.NewFakeRecorder(1),
		Client:      nil, // leave empty at this moment, test cases should instantiate one if Client is needed
		Helper:      nil, // leave empty at this moment, test cases should instantiate one if Helper is needed
		runtimeInfo: runtimeInfo,
	}

	return engine
}
