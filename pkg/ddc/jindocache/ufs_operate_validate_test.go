package jindocache

import (
	"context"
	"os"
	"reflect"

	. "github.com/agiledragon/gomonkey/v2"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/jindocache/operations"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

type fakeOperation struct {
	object        ctrlclient.Object
	operationType dataoperation.OperationType
}

func (f fakeOperation) HasPrecedingOperation() bool                                     { return false }
func (f fakeOperation) GetOperationObject() ctrlclient.Object                           { return f.object }
func (f fakeOperation) GetPossibleTargetDatasetNamespacedNames() []types.NamespacedName { return nil }
func (f fakeOperation) GetTargetDataset() (*datav1alpha1.Dataset, error)                { return nil, nil }
func (f fakeOperation) GetReleaseNameSpacedName() types.NamespacedName                  { return types.NamespacedName{} }
func (f fakeOperation) GetChartsDirectory() string                                      { return "" }
func (f fakeOperation) GetOperationType() dataoperation.OperationType                   { return f.operationType }
func (f fakeOperation) UpdateOperationApiStatus(*datav1alpha1.OperationStatus) error    { return nil }
func (f fakeOperation) Validate(cruntime.ReconcileRequestContext) ([]datav1alpha1.Condition, error) {
	return nil, nil
}
func (f fakeOperation) UpdateStatusInfoForCompleted(map[string]string) error      { return nil }
func (f fakeOperation) SetTargetDatasetStatusInProgress(*datav1alpha1.Dataset)    {}
func (f fakeOperation) RemoveTargetDatasetStatusInProgress(*datav1alpha1.Dataset) {}
func (f fakeOperation) GetStatusHandler() dataoperation.StatusHandler             { return nil }
func (f fakeOperation) GetTTL() (*int32, error)                                   { return nil, nil }
func (f fakeOperation) GetParallelTaskNumber() int32                              { return 1 }

var _ = Describe("JindoCacheEngine UFS, operation and validate helpers", func() {
	newEngine := func(name string, objects ...runtime.Object) *JindoCacheEngine {
		fakeClient := fake.NewFakeClientWithScheme(testScheme, objects...)
		runtime := &datav1alpha1.JindoRuntime{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "fluid"},
			Spec:       datav1alpha1.JindoRuntimeSpec{},
		}
		engine := &JindoCacheEngine{
			runtime:     runtime,
			name:        name,
			namespace:   "fluid",
			engineImpl:  common.JindoRuntime,
			Log:         fake.NullLogger(),
			Client:      fakeClient,
			runtimeInfo: buildRuntimeInfoForCleanAll(name),
		}
		return engine
	}

	It("should reject unsupported data operation types", func() {
		engine := newEngine("unsupported-op")
		operation := fakeOperation{
			operationType: dataoperation.DataBackupType,
			object: &datav1alpha1.DataProcess{
				TypeMeta: metav1.TypeMeta{APIVersion: datav1alpha1.GroupVersion.String(), Kind: "DataProcess"},
			},
		}

		_, err := engine.GetDataOperationValueFile(cruntime.ReconcileRequestContext{}, operation)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("not supported"))
	})

	It("should validate once runtime info has owner and placement data", func() {
		runtimeObject := &datav1alpha1.JindoRuntime{ObjectMeta: metav1.ObjectMeta{Name: "validate", Namespace: "fluid"}}
		engine := newEngine("validate", runtimeObject)
		uid := types.UID("dataset-uid")
		runtimeInfo, err := engine.getRuntimeInfo()
		Expect(err).NotTo(HaveOccurred())
		runtimeInfo.SetOwnerDatasetUID(uid)
		runtimeInfo.SetupWithDataset(&datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{Name: "validate", Namespace: "fluid", UID: uid},
			Spec:       datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ExclusiveMode},
		})

		Expect(engine.Validate(cruntime.ReconcileRequestContext{})).To(Succeed())
	})

	It("should return a type error for non-DataProcess objects", func() {
		engine := newEngine("process-type")

		_, err := engine.generateDataProcessValueFile(cruntime.ReconcileRequestContext{}, &corev1.ConfigMap{})

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("is not of type DataProcess"))
	})

	It("should generate a value file for a valid DataProcess", func() {
		dataset := &datav1alpha1.Dataset{ObjectMeta: metav1.ObjectMeta{Name: "dataset", Namespace: "fluid"}}
		dataProcess := &datav1alpha1.DataProcess{
			ObjectMeta: metav1.ObjectMeta{Name: "process", Namespace: "fluid"},
			Spec: datav1alpha1.DataProcessSpec{
				Dataset: datav1alpha1.TargetDatasetWithMountPath{
					TargetDataset: datav1alpha1.TargetDataset{Name: "dataset", Namespace: "fluid"},
					MountPath:     "/data",
				},
				Processor: datav1alpha1.Processor{
					Script: &datav1alpha1.ScriptProcessor{
						VersionSpec: datav1alpha1.VersionSpec{Image: "busybox"},
						Source:      "echo hello",
					},
				},
			},
		}
		engine := newEngine("dataset", dataset, dataProcess)

		valueFile, err := engine.generateDataProcessValueFile(cruntime.ReconcileRequestContext{}, dataProcess)

		Expect(err).NotTo(HaveOccurred())
		Expect(valueFile).NotTo(BeEmpty())
		defer func() { _ = os.Remove(valueFile) }()
		_, statErr := os.Stat(valueFile)
		Expect(statErr).NotTo(HaveOccurred())
	})

	It("should return an error when PrepareUFS runs before runtime is ready", func() {
		engine := newEngine("prepare-not-ready")

		err := engine.PrepareUFS()

		Expect(err).To(MatchError("runtime engine is not ready"))
	})

	It("should stop PrepareUFS when shouldMountUFS fails", func() {
		engine := newEngine("prepare-mount-error")

		readyPatch := ApplyMethod(reflect.TypeOf(engine), "CheckRuntimeReady", func(_ *JindoCacheEngine) bool {
			return true
		})
		defer readyPatch.Reset()
		mountCheckPatch := ApplyPrivateMethod(engine, "shouldMountUFS", func() (bool, error) {
			return false, context.DeadlineExceeded
		})
		defer mountCheckPatch.Reset()

		err := engine.PrepareUFS()

		Expect(err).To(MatchError(context.DeadlineExceeded))
	})

	It("should ignore SyncMetadata errors after successful UFS preparation", func() {
		engine := newEngine("prepare-sync-metadata")

		readyPatch := ApplyMethod(reflect.TypeOf(engine), "CheckRuntimeReady", func(_ *JindoCacheEngine) bool {
			return true
		})
		defer readyPatch.Reset()
		mountCheckPatch := ApplyPrivateMethod(engine, "shouldMountUFS", func() (bool, error) {
			return true, nil
		})
		defer mountCheckPatch.Reset()
		mountPatch := ApplyPrivateMethod(engine, "mountUFS", func() error {
			return nil
		})
		defer mountPatch.Reset()
		refreshCheckPatch := ApplyMethod(reflect.TypeOf(engine), "ShouldRefreshCacheSet", func(_ *JindoCacheEngine) (bool, error) {
			return true, nil
		})
		defer refreshCheckPatch.Reset()
		refreshPatch := ApplyMethod(reflect.TypeOf(engine), "RefreshCacheSet", func(_ *JindoCacheEngine) error {
			return nil
		})
		defer refreshPatch.Reset()
		syncMetadataPatch := ApplyMethod(reflect.TypeOf(engine), "SyncMetadata", func(_ *JindoCacheEngine) error {
			return context.Canceled
		})
		defer syncMetadataPatch.Reset()

		Expect(engine.PrepareUFS()).To(Succeed())
	})

	It("should re-sync dataset mounts when master starts after recorded mount time", func() {
		mountTime := metav1.NewTime(metav1.Now().Add(-time.Hour))
		runtime := &datav1alpha1.JindoRuntime{
			ObjectMeta: metav1.ObjectMeta{Name: "sync-mounts", Namespace: "fluid"},
			Spec:       datav1alpha1.JindoRuntimeSpec{},
			Status:     datav1alpha1.RuntimeStatus{MountTime: &mountTime},
		}
		masterPod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "sync-mounts-jindofs-master-0", Namespace: "fluid"},
			Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{{
				Name:  runtimeFSType + "-master",
				State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{StartedAt: metav1.Now()}},
			}}},
		}
		engine := newEngine("sync-mounts", runtime, masterPod)

		should, err := engine.ShouldSyncDatasetMounts()

		Expect(err).NotTo(HaveOccurred())
		Expect(should).To(BeTrue())
	})

	It("should not sync dataset mounts when the master is disabled", func() {
		runtime := &datav1alpha1.JindoRuntime{
			ObjectMeta: metav1.ObjectMeta{Name: "sync-disabled", Namespace: "fluid"},
			Spec:       datav1alpha1.JindoRuntimeSpec{Master: datav1alpha1.JindoCompTemplateSpec{Disabled: true}},
		}
		engine := newEngine("sync-disabled", runtime)

		should, err := engine.ShouldSyncDatasetMounts()

		Expect(err).NotTo(HaveOccurred())
		Expect(should).To(BeFalse())
	})

	It("should return an error when the runtime cannot be loaded before syncing dataset mounts", func() {
		engine := newEngine("missing-runtime")

		should, err := engine.ShouldSyncDatasetMounts()

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to get runtime when checking if dataset mounts need to be synced"))
		Expect(should).To(BeFalse())
	})

	It("should skip syncing dataset mounts when the master pod does not exist", func() {
		runtime := &datav1alpha1.JindoRuntime{ObjectMeta: metav1.ObjectMeta{Name: "missing-master-pod", Namespace: "fluid"}}
		engine := newEngine("missing-master-pod", runtime)

		should, err := engine.ShouldSyncDatasetMounts()

		Expect(err).NotTo(HaveOccurred())
		Expect(should).To(BeFalse())
	})

	It("should re-sync dataset mounts when mount time has not been recorded yet", func() {
		runtime := &datav1alpha1.JindoRuntime{ObjectMeta: metav1.ObjectMeta{Name: "sync-without-mount-time", Namespace: "fluid"}}
		masterPod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "sync-without-mount-time-jindofs-master-0", Namespace: "fluid"},
			Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{{
				Name:  runtimeFSType + "-master",
				State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{StartedAt: metav1.Now()}},
			}}},
		}
		engine := newEngine("sync-without-mount-time", runtime, masterPod)

		should, err := engine.ShouldSyncDatasetMounts()

		Expect(err).NotTo(HaveOccurred())
		Expect(should).To(BeTrue())
	})

	It("should wait when the master container is not running yet", func() {
		mountTime := metav1.Now()
		runtime := &datav1alpha1.JindoRuntime{
			ObjectMeta: metav1.ObjectMeta{Name: "sync-master-not-running", Namespace: "fluid"},
			Status:     datav1alpha1.RuntimeStatus{MountTime: &mountTime},
		}
		masterPod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "sync-master-not-running-jindofs-master-0", Namespace: "fluid"},
			Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{{
				Name:  runtimeFSType + "-master",
				State: corev1.ContainerState{},
			}}},
		}
		engine := newEngine("sync-master-not-running", runtime, masterPod)

		should, err := engine.ShouldSyncDatasetMounts()

		Expect(err).NotTo(HaveOccurred())
		Expect(should).To(BeFalse())
	})

	It("should keep dataset mounts in sync when the master container status is missing", func() {
		mountTime := metav1.Now()
		runtime := &datav1alpha1.JindoRuntime{
			ObjectMeta: metav1.ObjectMeta{Name: "sync-master-container-missing", Namespace: "fluid"},
			Status:     datav1alpha1.RuntimeStatus{MountTime: &mountTime},
		}
		masterPod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "sync-master-container-missing-jindofs-master-0", Namespace: "fluid"},
			Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{{
				Name:  "sidecar",
				State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{StartedAt: metav1.Now()}},
			}}},
		}
		engine := newEngine("sync-master-container-missing", runtime, masterPod)

		should, err := engine.ShouldSyncDatasetMounts()

		Expect(err).NotTo(HaveOccurred())
		Expect(should).To(BeFalse())
	})

	It("should aggregate total Jindo storage bytes from mounted paths", func() {
		dataset := &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{Name: "storage", Namespace: "fluid"},
			Spec: datav1alpha1.DatasetSpec{Mounts: []datav1alpha1.Mount{
				{Name: "bucket-a", MountPoint: "oss://bucket-a"},
				{Name: "bucket-b", MountPoint: "oss://bucket-b"},
			}},
		}
		engine := newEngine("storage", dataset)

		patch := ApplyPrivateMethod(operations.JindoFileUtils{}, "exec", func(_ operations.JindoFileUtils, command []string, verbose bool) (string, string, error) {
			switch command[len(command)-1] {
			case "jindo:///bucket-a":
				return "0 0 10 jindo:///bucket-a", "", nil
			case "jindo:///bucket-b":
				return "0 0 20 jindo:///bucket-b", "", nil
			default:
				return "0 0 0 unknown", "", nil
			}
		})
		defer patch.Reset()

		total, err := engine.TotalJindoStorageBytes()

		Expect(err).NotTo(HaveOccurred())
		Expect(total).To(Equal(int64(30)))
	})

	It("should keep unsupported UFS update and storage operations as no-ops", func() {
		engine := newEngine("ufs-no-op")

		shouldCheck, err := engine.ShouldCheckUFS()
		Expect(err).NotTo(HaveOccurred())
		Expect(shouldCheck).To(BeTrue())

		used, err := engine.UsedStorageBytes()
		Expect(err).NotTo(HaveOccurred())
		Expect(used).To(BeZero())

		free, err := engine.FreeStorageBytes()
		Expect(err).NotTo(HaveOccurred())
		Expect(free).To(BeZero())

		total, err := engine.TotalStorageBytes()
		Expect(err).NotTo(HaveOccurred())
		Expect(total).To(BeZero())

		files, err := engine.TotalFileNums()
		Expect(err).NotTo(HaveOccurred())
		Expect(files).To(BeZero())

		Expect(engine.ShouldUpdateUFS()).To(BeNil())

		updated, err := engine.UpdateOnUFSChange(nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(updated).To(BeFalse())
	})

	It("should return the jindo report summary from the master helper", func() {
		engine := newEngine("report-summary")

		patch := ApplyMethod(reflect.TypeOf(operations.JindoFileUtils{}), "ReportSummary", func(_ operations.JindoFileUtils) (string, error) {
			return "total cached: 10Gi", nil
		})
		defer patch.Reset()

		summary, err := engine.GetReportSummary()

		Expect(err).NotTo(HaveOccurred())
		Expect(summary).To(Equal("total cached: 10Gi"))
	})

	It("should update runtime mount time after syncing dataset mounts", func() {
		runtime := &datav1alpha1.JindoRuntime{ObjectMeta: metav1.ObjectMeta{Name: "sync-runtime-status", Namespace: "fluid"}}
		engine := newEngine("sync-runtime-status", runtime)

		preparePatch := ApplyMethod(reflect.TypeOf(engine), "PrepareUFS", func(_ *JindoCacheEngine) error {
			return nil
		})
		defer preparePatch.Reset()

		Expect(engine.SyncDatasetMounts()).To(Succeed())

		updatedRuntime := &datav1alpha1.JindoRuntime{}
		Expect(engine.Client.Get(context.TODO(), types.NamespacedName{Name: runtime.Name, Namespace: runtime.Namespace}, updatedRuntime)).To(Succeed())
		Expect(updatedRuntime.Status.MountTime).NotTo(BeNil())
	})

	It("should return an error when checking whether a UFS mount exists fails", func() {
		dataset := &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{Name: "mount-check-error", Namespace: "fluid"},
			Spec: datav1alpha1.DatasetSpec{Mounts: []datav1alpha1.Mount{{
				Name:       "bucket-a",
				MountPoint: "oss://bucket-a",
			}}},
		}
		engine := newEngine("mount-check-error", dataset)

		patch := ApplyMethod(reflect.TypeOf(operations.JindoFileUtils{}), "IsMounted", func(_ operations.JindoFileUtils, mountPoint string) (bool, error) {
			Expect(mountPoint).To(Equal("/bucket-a"))
			return false, context.DeadlineExceeded
		})
		defer patch.Reset()

		should, err := engine.shouldMountUFS()

		Expect(err).To(MatchError(context.DeadlineExceeded))
		Expect(should).To(BeFalse())
	})

	It("should request a UFS mount when one dataset mount is not mounted yet", func() {
		dataset := &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{Name: "mount-needed", Namespace: "fluid"},
			Spec: datav1alpha1.DatasetSpec{Mounts: []datav1alpha1.Mount{{
				Name:       "bucket-a",
				MountPoint: "oss://bucket-a",
			}}},
		}
		engine := newEngine("mount-needed", dataset)

		patch := ApplyMethod(reflect.TypeOf(operations.JindoFileUtils{}), "IsMounted", func(_ operations.JindoFileUtils, mountPoint string) (bool, error) {
			Expect(mountPoint).To(Equal("/bucket-a"))
			return false, nil
		})
		defer patch.Reset()

		should, err := engine.shouldMountUFS()

		Expect(err).NotTo(HaveOccurred())
		Expect(should).To(BeTrue())
	})

	It("should mount dataset UFS paths including pvc-backed local storage", func() {
		dataset := &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{Name: "mount-ufs", Namespace: "fluid"},
			Spec: datav1alpha1.DatasetSpec{Mounts: []datav1alpha1.Mount{
				{Name: "bucket-a", MountPoint: "oss://bucket-a"},
				{Name: "local-pvc", MountPoint: "pvc://demo-pvc/subdir"},
			}},
		}
		engine := newEngine("mount-ufs", dataset)
		calls := map[string]string{}

		patch := ApplyMethod(reflect.TypeOf(operations.JindoFileUtils{}), "Mount", func(_ operations.JindoFileUtils, mountPathInJindo string, ufsPath string) error {
			calls[mountPathInJindo] = ufsPath
			return nil
		})
		defer patch.Reset()

		Expect(engine.mountUFS()).To(Succeed())
		Expect(calls).To(HaveLen(2))
		Expect(calls).To(HaveKeyWithValue("/bucket-a", "oss://bucket-a"))
		Expect(calls).To(HaveKeyWithValue("/local-pvc", "local:///underFSStorage/local-pvc"))
	})

	It("should return an error when refreshing cache sets fails", func() {
		engine := newEngine("refresh-error")

		patch := ApplyMethod(reflect.TypeOf(operations.JindoFileUtils{}), "IsRefreshed", func(_ operations.JindoFileUtils) (bool, error) {
			return false, context.Canceled
		})
		defer patch.Reset()

		shouldRefresh, err := engine.ShouldRefreshCacheSet()

		Expect(err).To(MatchError(context.Canceled))
		Expect(shouldRefresh).To(BeFalse())
	})

	It("should refresh cache sets when jindocache has not loaded cachesets yet", func() {
		engine := newEngine("refresh-needed")
		refreshCalls := 0

		isRefreshedPatch := ApplyMethod(reflect.TypeOf(operations.JindoFileUtils{}), "IsRefreshed", func(_ operations.JindoFileUtils) (bool, error) {
			return false, nil
		})
		defer isRefreshedPatch.Reset()
		refreshPatch := ApplyMethod(reflect.TypeOf(operations.JindoFileUtils{}), "RefreshCacheSet", func(_ operations.JindoFileUtils) error {
			refreshCalls++
			return nil
		})
		defer refreshPatch.Reset()

		shouldRefresh, err := engine.ShouldRefreshCacheSet()
		Expect(err).NotTo(HaveOccurred())
		Expect(shouldRefresh).To(BeTrue())

		Expect(engine.RefreshCacheSet()).To(Succeed())
		Expect(refreshCalls).To(Equal(1))
	})
})

var _ = Describe("schema helpers", func() {
	It("keeps fake operations typed", func() {
		obj := &datav1alpha1.DataProcess{TypeMeta: metav1.TypeMeta{APIVersion: datav1alpha1.GroupVersion.String(), Kind: "DataProcess"}}
		Expect(obj.GetObjectKind().GroupVersionKind()).To(Equal(schema.GroupVersionKind{Group: datav1alpha1.GroupVersion.Group, Version: datav1alpha1.GroupVersion.Version, Kind: "DataProcess"}))
	})
})
