package alluxio

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
	MasterSts  *appsv1.StatefulSet
	WorkerSts  *appsv1.StatefulSet
	FuseDs     *appsv1.DaemonSet
	ConfigMaps map[string]*corev1.ConfigMap
	Services   map[string]*corev1.Service

	PersistentVolume      *corev1.PersistentVolume
	PersistentVolumeClaim *corev1.PersistentVolumeClaim
}

func mockFluidObjectsForTests(namespacedName types.NamespacedName) (*datav1alpha1.Dataset, *datav1alpha1.AlluxioRuntime) {
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

	tieredStoreQuota := resource.MustParse("2Gi")
	runtimeUid := uuid.NewUUID()
	runtime := &datav1alpha1.AlluxioRuntime{
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
		Spec: datav1alpha1.AlluxioRuntimeSpec{
			Replicas: 2,
			TieredStore: datav1alpha1.TieredStore{
				Levels: []datav1alpha1.Level{
					{
						MediumType: "MEM",
						Path:       "/dev/shm",
						Quota:      &tieredStoreQuota,
						High:       "0.95",
						Low:        "0.7",
					},
				},
			},
		},
	}

	return dataset, runtime
}

func mockAlluxioEngineForTests(dataset *datav1alpha1.Dataset, runtime *datav1alpha1.AlluxioRuntime) *AlluxioEngine {
	if dataset.Namespace != runtime.Namespace ||
		dataset.Name != runtime.Name ||
		len(runtime.OwnerReferences) == 0 ||
		runtime.OwnerReferences[0].UID != dataset.UID {

		panic("engine mocking is only valid on a valid pair of dataset and runtime")
	}

	// build a very basic runtime info here. Test cases can override it with a new runtimeInfo.
	runtimeInfo, _ := base.BuildRuntimeInfo(dataset.Name, dataset.Namespace, common.AlluxioRuntime)

	engine := &AlluxioEngine{
		runtime:     runtime,
		name:        runtime.Name,
		namespace:   runtime.Namespace,
		runtimeType: common.AlluxioRuntime,
		engineImpl:  common.AlluxioEngineImpl,
		Log:         fake.NullLogger(),
		Recorder:    record.NewFakeRecorder(1),
		Client:      nil, // leave empty at this moment, test cases should handle this if Client is needed
		Helper:      nil, // leave empty at this moment, test cases should handle this if Helper is needed
		runtimeInfo: runtimeInfo,
		UnitTest:    true,
	}

	return engine
}

func mockAlluxioObjectsForTests(dataset *datav1alpha1.Dataset, runtime *datav1alpha1.AlluxioRuntime, engine *AlluxioEngine) mockedObjects {
	if dataset.Namespace != runtime.Namespace ||
		dataset.Name != runtime.Name ||
		len(runtime.OwnerReferences) == 0 ||
		runtime.OwnerReferences[0].UID != dataset.UID {

		panic("alluxio objects mocking is only valid on a valid pair of dataset and runtime")
	}

	commonLabels := map[string]string{
		"app":                          "alluxio",
		"chart":                        "alluxio-0.9.13",
		"release":                      runtime.Name,
		"heritage":                     "Helm",
		"app.kubernetes.io/name":       "alluxio",
		"app.kubernetes.io/instance":   runtime.Name,
		"app.kubernetes.io/managed-by": "Helm",
		"fluid.io/managed-by":          "fluid",
		"fluid.io/dataset-id":          utils.GetDatasetId(dataset.Namespace, dataset.Name, string(dataset.UID)),
	}

	// mock master sts
	masterLabels := map[string]string{
		"role": "alluxio-master",
		"name": engine.getMasterName(),
	}
	for k, v := range commonLabels {
		masterLabels[k] = v
	}
	masterSts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      engine.getMasterName(),
			Namespace: runtime.Namespace,
			Labels:    masterLabels,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: datav1alpha1.GroupVersion.String(),
					Kind:       "Alluxio",
					Name:       runtime.Name,
					UID:        runtime.UID,
				},
			},
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: fmt.Sprintf("%s-master", runtime.Name),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":  masterLabels["app"],
					"role": masterLabels["role"],
					"name": masterLabels["name"],
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: masterLabels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "alluxio-master",
						},
						{
							Name: "alluxio-job-master",
						},
					},
				},
			},
		},
	}

	// mock worker sts
	workerLabels := map[string]string{
		"role":             "alluxio-worker",
		"name":             engine.getWorkerName(),
		"fluid.io/dataset": utils.GetDatasetId(dataset.Namespace, dataset.Name, string(dataset.UID)),
		// "fluid.io/dataset-placement": string(dataset.Spec.PlacementMode),
	}
	for k, v := range commonLabels {
		workerLabels[k] = v
	}
	workerSts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      engine.getWorkerName(),
			Namespace: runtime.Namespace,
			Labels:    workerLabels,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: datav1alpha1.GroupVersion.String(),
					Kind:       "Alluxio",
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
							Name: "alluxio-worker",
						},
						{
							Name: "alluxio-job-worker",
						},
					},
				},
			},
		},
	}

	// mock fuse ds
	fuseLabels := map[string]string{
		"role": "alluxio-fuse",
		"name": engine.getFuseName(),
	}
	for k, v := range commonLabels {
		fuseLabels[k] = v
	}

	fuseDs := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      engine.getFuseName(),
			Namespace: runtime.Namespace,
			Labels:    fuseLabels,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: datav1alpha1.GroupVersion.String(),
					Kind:       "Alluxio",
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
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: fuseLabels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "alluxio-fuse",
						},
					},
				},
			},
		},
	}

	// mock master service
	masterCnt := 1
	if runtime.Spec.Master.Replicas > 0 {
		masterCnt = int(runtime.Spec.Master.Replicas)
	}

	masterServices := map[string]*corev1.Service{}
	serviceLabels := map[string]string{
		"role": "alluxio-master",
	}
	for k, v := range commonLabels {
		serviceLabels[k] = v
	}
	for i := 0; i < masterCnt; i++ {
		svcName := fmt.Sprintf("%s-master-%d", runtime.Name, i)
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      svcName,
				Namespace: runtime.Namespace,
				Labels:    serviceLabels,
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: datav1alpha1.GroupVersion.String(),
						Kind:       "Alluxio",
						Name:       runtime.Name,
						UID:        runtime.UID,
					},
				},
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{
					"role":                               serviceLabels["role"],
					"app":                                serviceLabels["app"],
					"release":                            serviceLabels["release"],
					"statefulset.kubernetes.io/pod-name": svcName,
				},
			},
		}
		masterServices[svcName] = svc
	}

	// mock configmaps
	cmLabels := map[string]string{
		"name": fmt.Sprintf("%s-config", runtime.Name),
	}
	for k, v := range commonLabels {
		cmLabels[k] = v
	}
	configMaps := map[string]*corev1.ConfigMap{}
	configMaps[fmt.Sprintf("%s-config", runtime.Name)] = &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-config", runtime.Name),
			Namespace: runtime.Namespace,
			Labels:    cmLabels,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: datav1alpha1.GroupVersion.String(),
					Kind:       "Alluxio",
					Name:       runtime.Name,
					UID:        runtime.UID,
				},
			},
		},
		Data: map[string]string{},
	}

	mountCmLabels := map[string]string{
		"name": fmt.Sprintf("%s-mount-config", runtime.Name),
	}
	for k, v := range commonLabels {
		mountCmLabels[k] = v
	}
	configMaps[fmt.Sprintf("%s-mount-config", runtime.Name)] = &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-mount-config", runtime.Name),
			Namespace: runtime.Namespace,
			Labels:    mountCmLabels,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: datav1alpha1.GroupVersion.String(),
					Kind:       "Alluxio",
					Name:       runtime.Name,
					UID:        runtime.UID,
				},
			},
		},
		Data: map[string]string{},
	}

	// mock persistent volume
	accessModes := []corev1.PersistentVolumeAccessMode{corev1.ReadOnlyMany}
	if len(dataset.Spec.AccessModes) != 0 {
		accessModes = dataset.Spec.AccessModes
	}
	mountPath := engine.getMountPoint()
	pv := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", runtime.Namespace, runtime.Name),
			Namespace: runtime.Namespace,
			Labels: map[string]string{
				common.LabelAnnotationDatasetId: utils.GetDatasetId(runtime.Namespace, runtime.Name, string(dataset.UID)),
				utils.GetCommonLabelName(false, runtime.Namespace, runtime.Name, string(dataset.UID)): "true",
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
						common.VolumeAttrMountType: common.AlluxioMountType,
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
				utils.GetCommonLabelName(false, runtime.Namespace, runtime.Name, string(dataset.UID)): "true",
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
		MasterSts:  masterSts,
		WorkerSts:  workerSts,
		FuseDs:     fuseDs,
		Services:   masterServices,
		ConfigMaps: configMaps,

		PersistentVolume:      pv,
		PersistentVolumeClaim: pvc,
	}
}

// mock result from executing command [alluxio fsadmin report summary]
func mockAlluxioReportSummary(used, capacity string) string {
	// e.g. used = 0B, capacity = 19.07MB
	s := `Alluxio cluster summary: 
	Master Address: 172.18.0.2:20000
	Web Port: 20001
	Rpc Port: 20000
	Started: 06-29-2021 13:43:56:297
	Uptime: 0 day(s), 0 hour(s), 4 minute(s), and 13 second(s)
	Version: 2.3.1-SNAPSHOT
	Safe Mode: false
	Zookeeper Enabled: false
	Live Workers: 1
	Lost Workers: 0
	Total Capacity: %[1]s
		Tier: MEM  Size: %[1]s
	Used Capacity: %[2]s
		Tier: MEM  Size: %[2]s
	Free Capacity: <fluid-do-not-care-this-value>
	`
	return fmt.Sprintf(s, capacity, used)
}

func mockAlluxioReportMetrics(bytesReadLocal, bytesReadLocalThroughput, bytesReadRemote, bytesReadRemoteThroughput, bytesReadUfs, bytesReadUfsThroughput string) string {
	// e.g. bytesReadLocal = 19.37MB, bytesReadLocalThroughput = 495.97KB/MIN, bytesReadUfs = 30.75MB, bytesReadUfsThroughput = 787.17KB/MIN
	r := `Cluster.BytesReadAlluxio  (Type: COUNTER, Value: 0B)
	Cluster.BytesReadAlluxioThroughput  (Type: GAUGE, Value: 0B/MIN)
	Cluster.BytesReadDomain  (Type: COUNTER, Value: 0B)
	Cluster.BytesReadDomainThroughput  (Type: GAUGE, Value: 0B/MIN)
	Cluster.BytesReadLocal  (Type: COUNTER, Value: %[1]s)
	Cluster.BytesReadLocalThroughput  (Type: GAUGE, Value: %[2]s)
	Cluster.BytesReadRemote (Type: COUNTER, Value: %[3]s)
	Cluster.BytesReadRemoteThroughput  (Type: GAUGE, Value: %[4]s)
	Cluster.BytesReadPerUfs.UFS:s3:fluid  (Type: COUNTER, Value: %[5]s)
	Cluster.BytesReadUfsAll  (Type: COUNTER, Value: %[5]s)
	Cluster.BytesReadUfsThroughput  (Type: GAUGE, Value: %[6]s)
	Cluster.BytesWrittenAlluxio  (Type: COUNTER, Value: 0B)
	Cluster.BytesWrittenAlluxioThroughput  (Type: GAUGE, Value: 0B/MIN)
	Cluster.BytesWrittenDomain  (Type: COUNTER, Value: 0B)
	Cluster.BytesWrittenDomainThroughput  (Type: GAUGE, Value: 0B/MIN)
	Cluster.BytesWrittenLocal  (Type: COUNTER, Value: 0B)
	Cluster.BytesWrittenLocalThroughput  (Type: GAUGE, Value: 0B/MIN)
	Cluster.BytesWrittenUfsAll  (Type: COUNTER, Value: 0B)
	Cluster.BytesWrittenUfsThroughput  (Type: GAUGE, Value: 0B/MIN)
	Cluster.CapacityFree  (Type: GAUGE, Value: 9,842,601)
	Cluster.CapacityFreeTierHDD  (Type: GAUGE, Value: 0)
	Cluster.CapacityFreeTierMEM  (Type: GAUGE, Value: 9,842,601)
	Cluster.CapacityFreeTierSSD  (Type: GAUGE, Value: 0)
	Cluster.CapacityTotal  (Type: GAUGE, Value: 20,000,000)
	Cluster.CapacityTotalTierHDD  (Type: GAUGE, Value: 0)
	Cluster.CapacityTotalTierMEM  (Type: GAUGE, Value: 20,000,000)
	Cluster.CapacityTotalTierSSD  (Type: GAUGE, Value: 0)
	Cluster.CapacityUsed  (Type: GAUGE, Value: 10,157,399)
	Cluster.CapacityUsedTierHDD  (Type: GAUGE, Value: 0)
	Cluster.CapacityUsedTierMEM  (Type: GAUGE, Value: 10,157,399)
	Cluster.CapacityUsedTierSSD  (Type: GAUGE, Value: 0)
	Cluster.RootUfsCapacityFree  (Type: GAUGE, Value: -1)
	Cluster.RootUfsCapacityTotal  (Type: GAUGE, Value: -1)
	Cluster.RootUfsCapacityUsed  (Type: GAUGE, Value: -1)
	Cluster.Workers  (Type: GAUGE, Value: 1)
	Master.CompleteFileOps  (Type: COUNTER, Value: 0)
	Master.ConnectFromMaster.UFS:s3:%2F%2Ffluid.UFS_TYPE:s3  (Type: TIMER, Value: 0)
	Master.Create.UFS:%2Fjournal%2FBlockMaster.UFS_TYPE:local  (Type: TIMER, Value: 1)
	Master.Create.UFS:%2Fjournal%2FFileSystemMaster.UFS_TYPE:local  (Type: TIMER, Value: 1)
	Master.Create.UFS:%2Fjournal%2FMetaMaster.UFS_TYPE:local  (Type: TIMER, Value: 1)
	Master.CreateDirectoryOps  (Type: COUNTER, Value: 0)
	Master.CreateFileOps  (Type: COUNTER, Value: 0)
	Master.DeletePathOps  (Type: COUNTER, Value: 0)
	Master.DirectoriesCreated  (Type: COUNTER, Value: 0)
	Master.EdgeCacheSize  (Type: GAUGE, Value: 7)
	Master.Exists.UFS:s3:%2F%2Ffluid.UFS_TYPE:s3  (Type: TIMER, Value: 2)
	Master.FileBlockInfosGot  (Type: COUNTER, Value: 0)
	Master.FileInfosGot  (Type: COUNTER, Value: 25)
	Master.FilesCompleted  (Type: COUNTER, Value: 7)
	Master.FilesCreated  (Type: COUNTER, Value: 7)
	Master.FilesFreed  (Type: COUNTER, Value: 0)
	Master.FilesPersisted  (Type: COUNTER, Value: 0)
	Master.FilesPinned  (Type: GAUGE, Value: 0)
	Master.FreeFileOps  (Type: COUNTER, Value: 0)
	Master.GetAcl.UFS:s3:%2F%2Ffluid.UFS_TYPE:s3  (Type: TIMER, Value: 7)
	Master.GetBlockInfo.User:root  (Type: TIMER, Value: 3)
	Master.GetBlockMasterInfo.User:root  (Type: TIMER, Value: 173)
	Master.GetConfigHash.User:root  (Type: TIMER, Value: 40)
	Master.GetFileBlockInfoOps  (Type: COUNTER, Value: 0)
	Master.GetFileInfoOps  (Type: COUNTER, Value: 9)
	Master.GetFileLocations.User:root.UFS:s3:%2F%2Ffluid.UFS_TYPE:s3  (Type: TIMER, Value: 24)
	Master.GetFingerprint.User:root.UFS:s3:%2F%2Ffluid.UFS_TYPE:s3  (Type: TIMER, Value: 1)
	Master.GetMountTable.User:root  (Type: TIMER, Value: 2)
	Master.GetNewBlockOps  (Type: COUNTER, Value: 0)
	Master.GetSpace.UFS:s3:%2F%2Ffluid.UFS_TYPE:s3  (Type: TIMER, Value: 18)
	Master.GetSpace.User:root.UFS:s3:%2F%2Ffluid.UFS_TYPE:s3  (Type: TIMER, Value: 103)
	Master.GetStatus.User:root  (Type: TIMER, Value: 6)
	Master.GetStatus.User:root.UFS:s3:%2F%2Ffluid.UFS_TYPE:s3  (Type: TIMER, Value: 3)
	Master.GetStatusFailures.User:root.UFS:s3:%2F%2Ffluid.UFS_TYPE:s3  (Type: COUNTER, Value: 2)
	Master.GetWorkerInfoList.User:root  (Type: TIMER, Value: 2)
	Master.InodeCacheSize  (Type: GAUGE, Value: 8)
	Master.JournalFlushTimer  (Type: TIMER, Value: 22)
	Master.LastBackupEntriesCount  (Type: GAUGE, Value: -1)
	Master.LastBackupRestoreCount  (Type: GAUGE, Value: -1)
	Master.LastBackupRestoreTimeMs  (Type: GAUGE, Value: -1)
	Master.LastBackupTimeMs  (Type: GAUGE, Value: -1)
	Master.ListStatus.UFS:%2Fjournal%2FBlockMaster.UFS_TYPE:local  (Type: TIMER, Value: 63)
	Master.ListStatus.UFS:%2Fjournal%2FFileSystemMaster.UFS_TYPE:local  (Type: TIMER, Value: 63)
	Master.ListStatus.UFS:%2Fjournal%2FMetaMaster.UFS_TYPE:local  (Type: TIMER, Value: 63)
	Master.ListStatus.UFS:%2Fjournal%2FMetricsMaster.UFS_TYPE:local  (Type: TIMER, Value: 63)
	Master.ListStatus.UFS:%2Fjournal%2FTableMaster.UFS_TYPE:local  (Type: TIMER, Value: 63)
	Master.ListStatus.User:root  (Type: TIMER, Value: 3)
	Master.ListStatus.User:root.UFS:s3:%2F%2Ffluid.UFS_TYPE:s3  (Type: TIMER, Value: 1)
	Master.ListingCacheSize  (Type: GAUGE, Value: 8)
	Master.MountOps  (Type: COUNTER, Value: 0)
	Master.NewBlocksGot  (Type: COUNTER, Value: 0)
	Master.PathsDeleted  (Type: COUNTER, Value: 0)
	Master.PathsMounted  (Type: COUNTER, Value: 0)
	Master.PathsRenamed  (Type: COUNTER, Value: 0)
	Master.PathsUnmounted  (Type: COUNTER, Value: 0)
	Master.PerUfsOpConnectFromMaster.UFS:s3:%2F%2Ffluid  (Type: GAUGE, Value: 0)
	Master.PerUfsOpCreate.UFS:%2Fjournal%2FBlockMaster  (Type: GAUGE, Value: 1)
	Master.PerUfsOpCreate.UFS:%2Fjournal%2FFileSystemMaster  (Type: GAUGE, Value: 1)
	Master.PerUfsOpCreate.UFS:%2Fjournal%2FMetaMaster  (Type: GAUGE, Value: 1)
	Master.PerUfsOpExists.UFS:s3:%2F%2Ffluid  (Type: GAUGE, Value: 2)
	Master.PerUfsOpGetFileLocations.UFS:s3:%2F%2Ffluid  (Type: GAUGE, Value: 24)
	Master.PerUfsOpGetFingerprint.UFS:s3:%2F%2Ffluid  (Type: GAUGE, Value: 1)
	Master.PerUfsOpGetSpace.UFS:s3:%2F%2Ffluid  (Type: GAUGE, Value: 116)
	Master.PerUfsOpGetStatus.UFS:s3:%2F%2Ffluid  (Type: GAUGE, Value: 3)
	Master.PerUfsOpListStatus.UFS:%2Fjournal%2FBlockMaster  (Type: GAUGE, Value: 60)
	Master.PerUfsOpListStatus.UFS:%2Fjournal%2FFileSystemMaster  (Type: GAUGE, Value: 60)
	Master.PerUfsOpListStatus.UFS:%2Fjournal%2FMetaMaster  (Type: GAUGE, Value: 60)
	Master.PerUfsOpListStatus.UFS:%2Fjournal%2FMetricsMaster  (Type: GAUGE, Value: 60)
	Master.PerUfsOpListStatus.UFS:%2Fjournal%2FTableMaster  (Type: GAUGE, Value: 60)
	Master.PerUfsOpListStatus.UFS:s3:%2F%2Ffluid  (Type: GAUGE, Value: 1)
	Master.RenamePathOps  (Type: COUNTER, Value: 0)
	Master.SetAclOps  (Type: COUNTER, Value: 0)
	Master.SetAttributeOps  (Type: COUNTER, Value: 0)
	Master.TotalPaths  (Type: GAUGE, Value: 8)
	Master.UfsSessionCount-Ufs:s3:%2F%2Ffluid  (Type: COUNTER, Value: 0)
	Master.UnmountOps  (Type: COUNTER, Value: 0)
	Master.blockHeartbeat.User:root  (Type: TIMER, Value: 2,410)
	Master.commitBlock.User:root  (Type: TIMER, Value: 1)
	Master.getConfigHash  (Type: TIMER, Value: 4)
	Master.getConfigHash.User:root  (Type: TIMER, Value: 239)
	Master.getConfiguration  (Type: TIMER, Value: 20)
	Master.getConfiguration.User:root  (Type: TIMER, Value: 428)
	Master.getMasterInfo.User:root  (Type: TIMER, Value: 173)
	Master.getMetrics.User:root  (Type: TIMER, Value: 33)
	Master.getPinnedFileIds.User:root  (Type: TIMER, Value: 2,410)
	Master.getUfsInfo.User:root  (Type: TIMER, Value: 1)
	Master.getWorkerId.User:root  (Type: TIMER, Value: 1)
	Master.metricsHeartbeat.User:root  (Type: TIMER, Value: 4)
	Master.registerWorker.User:root  (Type: TIMER, Value: 1)
	`
	return fmt.Sprintf(r, bytesReadLocal, bytesReadLocalThroughput, bytesReadRemote, bytesReadRemoteThroughput, bytesReadUfs, bytesReadUfsThroughput)
}
