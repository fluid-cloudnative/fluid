package thin

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/tools/record"
)

type mockedObjects struct {
	FuseDs     *appsv1.DaemonSet
	ConfigMaps map[string]*corev1.ConfigMap

	PersistentVolume      *corev1.PersistentVolume
	PersistentVolumeClaim *corev1.PersistentVolumeClaim
}

func mockFluidObjectsForTests(namespacedName types.NamespacedName) (*datav1alpha1.Dataset, *datav1alpha1.ThinRuntime, *datav1alpha1.ThinRuntimeProfile) {
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
					MountPoint: "s3://mybucket/mypath",
					Path:       "/",
					Name:       "s3-bucket",
					Options: map[string]string{
						"endpoint": "my-s3-endpoint",
					},
					EncryptOptions: []datav1alpha1.EncryptOption{
						{
							Name: "access-key-id",
							ValueFrom: datav1alpha1.EncryptOptionSource{
								SecretKeyRef: datav1alpha1.SecretKeySelector{
									Name: "my-s3-secret",
									Key:  "myak",
								},
							},
						},
						{
							Name: "access-key-secret",
							ValueFrom: datav1alpha1.EncryptOptionSource{
								SecretKeyRef: datav1alpha1.SecretKeySelector{
									Name: "my-s3-secret",
									Key:  "mysk",
								},
							},
						},
					},
				},
			},
		},
	}

	runtimeUid := uuid.NewUUID()
	runtime := &datav1alpha1.ThinRuntime{
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
		Spec: datav1alpha1.ThinRuntimeSpec{
			Replicas:               2,
			ThinRuntimeProfileName: "mock-profile",
		},
	}

	profileUid := uuid.NewUUID()
	profile := &datav1alpha1.ThinRuntimeProfile{
		ObjectMeta: metav1.ObjectMeta{
			Name: "mock-profile",
			UID:  profileUid,
		},
		Spec: datav1alpha1.ThinRuntimeProfileSpec{
			FileSystemType: "mock-fs",
			Fuse: datav1alpha1.ThinFuseSpec{
				Image:    "mock-image",
				ImageTag: "mock-tag",
			},
		},
	}

	return dataset, runtime, profile
}

func mockThinEngineForTests(dataset *datav1alpha1.Dataset, runtime *datav1alpha1.ThinRuntime, profile *datav1alpha1.ThinRuntimeProfile) *ThinEngine {
	if dataset.Namespace != runtime.Namespace ||
		dataset.Name != runtime.Name ||
		len(runtime.OwnerReferences) == 0 ||
		runtime.OwnerReferences[0].UID != dataset.UID {

		panic("engine mocking is only valid on a valid pair of dataset and runtime")
	}

	runtimeInfo, _ := base.BuildRuntimeInfo(dataset.Name, dataset.Namespace, common.ThinRuntime)

	engine := &ThinEngine{
		runtime:        runtime,
		runtimeProfile: profile,
		name:           runtime.Name,
		namespace:      runtime.Namespace,
		runtimeType:    common.ThinRuntime,
		engineImpl:     common.ThinEngineImpl,
		Log:            fake.NullLogger(),
		Recorder:       record.NewFakeRecorder(1),
		Client:         nil, // leave empty at this moment, test cases should handle this if Client is needed
		Helper:         nil, // leave empty at this moment, test cases should handle this if Helper is needed
		runtimeInfo:    runtimeInfo,
		UnitTest:       true,
	}

	return engine
}


