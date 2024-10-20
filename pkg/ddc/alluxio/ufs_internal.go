/*
Copyright 2020 The Fluid Author.

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

package alluxio

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/alluxio/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	securityutil "github.com/fluid-cloudnative/fluid/pkg/utils/security"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/util/retry"
)

func (e *AlluxioEngine) usedStorageBytesInternal() (value int64, err error) {
	return
}

func (e *AlluxioEngine) freeStorageBytesInternal() (value int64, err error) {
	return
}

func (e *AlluxioEngine) totalStorageBytesInternal() (total int64, err error) {
	podName, containerName := e.getMasterPodInfo()

	fileUitls := operations.NewAlluxioFileUtils(podName, containerName, e.namespace, e.Log)
	_, _, total, err = fileUitls.Count("/")
	if err != nil {
		return
	}

	return
}

func (e *AlluxioEngine) totalFileNumsInternal() (fileCount int64, err error) {
	podName, containerName := e.getMasterPodInfo()

	fileUitls := operations.NewAlluxioFileUtils(podName, containerName, e.namespace, e.Log)
	fileCount, err = fileUitls.GetFileCount()
	if err != nil {
		return
	}

	return
}

// shouldMountUFS checks if there's any UFS that need to be mounted
func (e *AlluxioEngine) shouldMountUFS() (should bool, err error) {
	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		return should, err
	}
	e.Log.Info("get dataset info", "dataset", dataset)

	podName, containerName := e.getMasterPodInfo()
	fileUtils := operations.NewAlluxioFileUtils(podName, containerName, e.namespace, e.Log)

	ready := fileUtils.Ready()
	if !ready {
		should = false
		err = fmt.Errorf("the UFS is not ready")
		return should, err
	}

	// Check if any of the Mounts has not been mounted in Alluxio
	for _, mount := range dataset.Spec.Mounts {
		if common.IsFluidNativeScheme(mount.MountPoint) {
			// No need for a mount point with Fluid native scheme('local://' and 'pvc://') to be mounted
			continue
		}
		alluxioPath := utils.UFSPathBuilder{}.GenUFSPathInUnifiedNamespace(mount)
		mounted, err := fileUtils.IsMounted(alluxioPath)
		if err != nil {
			should = false
			return should, err
		}
		if !mounted {
			e.Log.Info("Found dataset that is not mounted.", "dataset", dataset)
			should = true
			return should, err
		}
	}

	return should, err
}

// FindUnmountedUFS return if UFSs not mounted
func (e *AlluxioEngine) FindUnmountedUFS() (unmountedPaths []string, err error) {
	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	e.Log.Info("get dataset info", "dataset", dataset)
	if err != nil {
		return unmountedPaths, err
	}

	podName, containerName := e.getMasterPodInfo()
	fileUtils := operations.NewAlluxioFileUtils(podName, containerName, e.namespace, e.Log)

	ready := fileUtils.Ready()
	if !ready {
		err = fmt.Errorf("the UFS is not ready")
		return unmountedPaths, err
	}

	var alluxioPaths []string
	// Check if any of the Mounts has not been mounted in Alluxio
	for _, mount := range dataset.Spec.Mounts {
		if common.IsFluidNativeScheme(mount.MountPoint) {
			// No need for a mount point with Fluid native scheme('local://' and 'pvc://') to be mounted
			continue
		}
		alluxioPath := utils.UFSPathBuilder{}.GenUFSPathInUnifiedNamespace(mount)
		alluxioPaths = append(alluxioPaths, alluxioPath)
	}

	// For fluid native schema, skip mount check to avoid unnecessary system call
	if len(alluxioPaths) == 0 {
		return
	}

	return fileUtils.FindUnmountedAlluxioPaths(alluxioPaths)
}

func (e *AlluxioEngine) processUpdatingUFS(ufsToUpdate *utils.UFSToUpdate) (updateReady bool, err error) {
	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		return false, err
	}

	if IsMountWithConfigMap() {
		updateReady, err = e.updateUFSWithMountConfigMapScript(dataset)
	} else {
		updateReady, err = e.updatingUFSWithMountCommand(dataset, ufsToUpdate)
	}

	if err != nil {
		return
	}

	if updateReady {
		// need to reset ufsTotal to Calculating so that SyncMetadata will work
		datasetToUpdate := dataset.DeepCopy()
		datasetToUpdate.Status.UfsTotal = metadataSyncNotDoneMsg
		if !reflect.DeepEqual(dataset.Status, datasetToUpdate.Status) {
			err = e.Client.Status().Update(context.TODO(), datasetToUpdate)
			if err != nil {
				e.Log.Error(err, "fail to update ufsTotal of dataset to Calculating")
			}
		}

		err = e.SyncMetadata()
		if err != nil {
			// just report this error and ignore it because SyncMetadata isn't on the critical path of Setup
			e.Log.Error(err, "SyncMetadata", "dataset", e.name)
			return true, nil
		}

		// whether having new mount path alluxio
		if len(ufsToUpdate.ToAdd()) > 0 {
			e.updateMountTime()
		}
	}

	return
}

func (e *AlluxioEngine) updatingUFSWithMountCommand(dataset *datav1alpha1.Dataset, ufsToUpdate *utils.UFSToUpdate) (updateReady bool, err error) {

	podName, containerName := e.getMasterPodInfo()
	fileUtils := operations.NewAlluxioFileUtils(podName, containerName, e.namespace, e.Log)

	ready := fileUtils.Ready()
	if !ready {
		return false, fmt.Errorf("the UFS is not ready, namespace:%s,name:%s", e.namespace, e.name)
	}

	// Iterate all the mount points, do mount if the mount point is in added array
	// TODO: not allow to edit FluidNativeScheme MountPoint
	for _, mount := range dataset.Spec.Mounts {
		if common.IsFluidNativeScheme(mount.MountPoint) {
			continue
		}

		alluxioPath := utils.UFSPathBuilder{}.GenUFSPathInUnifiedNamespace(mount)
		if len(ufsToUpdate.ToAdd()) > 0 && utils.ContainsString(ufsToUpdate.ToAdd(), alluxioPath) {
			mountOptions := map[string]string{}
			for key, value := range dataset.Spec.SharedOptions {
				mountOptions[key] = value
			}

			for key, value := range mount.Options {
				mountOptions[key] = value
			}

			// Configure mountOptions using encryptOptions
			// If encryptOptions have the same key with options, it will overwrite the corresponding value
			mountOptions, err = e.genEncryptOptions(dataset.Spec.SharedEncryptOptions, mountOptions, mount.Name)
			if err != nil {
				return false, err
			}

			mountOptions, err = e.genEncryptOptions(mount.EncryptOptions, mountOptions, mount.Name)
			if err != nil {
				return false, err
			}

			err = fileUtils.Mount(alluxioPath, mount.MountPoint, mountOptions, mount.ReadOnly, mount.Shared)
			if err != nil {
				return false, err
			}
		}
	}

	// unmount the mount point in the removed array
	if len(ufsToUpdate.ToRemove()) > 0 {
		for _, mountRemove := range ufsToUpdate.ToRemove() {
			err = fileUtils.UnMount(mountRemove)
			if err != nil {
				return false, err
			}
		}
	}

	return true, nil
}

// update alluxio mount using script in configmap
func (e *AlluxioEngine) updateUFSWithMountConfigMapScript(dataset *datav1alpha1.Dataset) (updateReady bool, err error) {
	// 1. update non native mount info according the data.Spec.Mounts
	mountConfigMapName := e.getMountConfigmapName()
	mountConfigMap, err := kubeclient.GetConfigmapByName(e.Client, mountConfigMapName, e.namespace)
	if mountConfigMap == nil {
		// if configmap not found, considered as old dataset (runtime ClusterRole having secret resource)
		e.Recorder.Eventf(dataset, v1.EventTypeWarning, common.RuntimeWithSecretNotSupported,
			"dataset created by runtime without using configmap as mount storage is not support dynamic updates")
		return false, errors.Wrapf(errors.Errorf("mount configmap %s is not found", mountConfigMapName),
			"dataset %s may be created by runtime without using configmap as mount storage, "+
				"which does not support dynamic updating the mounts field.", dataset.GetName())
	}
	if err != nil {
		return false, err
	}

	nonNativeMounts, err := e.generateNonNativeMountsInfo(dataset)
	if err != nil {
		return false, err
	}

	newMountConfigMap := mountConfigMap.DeepCopy()
	newMountConfigMap.Data[NON_NATIVE_MOUNT_DATA_NAME] = strings.Join(nonNativeMounts, "\n") + "\n"
	if !reflect.DeepEqual(newMountConfigMap, mountConfigMap) {
		e.Log.Info("update mount path", NON_NATIVE_MOUNT_DATA_NAME, nonNativeMounts)
		err = kubeclient.UpdateConfigMap(e.Client, newMountConfigMap)
		if err != nil {
			return false, err
		}
	}

	// 2. execute mount script to mount and unmount alluxio path according to non native mount info
	podName, containerName := e.getMasterPodInfo()
	fileUtils := operations.NewAlluxioFileUtils(podName, containerName, e.namespace, e.Log)
	err = fileUtils.ExecMountScripts()
	if err != nil {
		return false, errors.Wrapf(err, "execute mount.sh occurs error")
	}

	// 3. compare alluxio mount paths and dataset mount path to mark updateReady because configmap updating has delay
	// for alluxio master pods so above executing may not use the newest non native mount info.
	mountedPaths, err := fileUtils.GetMountedAlluxioPaths()
	// the root path can not be changed, skip it.
	mountedPaths = utils.RemoveString(mountedPaths, "/")

	var datasetMountPaths []string
	for _, mount := range dataset.Spec.Mounts {
		if common.IsFluidNativeScheme(mount.MountPoint) {
			continue
		}
		m := utils.UFSPathBuilder{}.GenUFSPathInUnifiedNamespace(mount)
		// skip root path mount
		if m != "/" {
			datasetMountPaths = append(datasetMountPaths, m)
		}
	}
	// only ready when alluxio mounted paths is the same as dataset mount paths
	if len(mountedPaths) == len(datasetMountPaths) && len(utils.SubtractString(datasetMountPaths, mountedPaths)) == 0 {
		updateReady = true
	}

	return
}

// mountUFS() mount all UFSs to Alluxio according to mount points in `dataset.Spec`. If a mount point is Fluid-native, mountUFS() will skip it.
func (e *AlluxioEngine) mountUFS() (err error) {
	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		return err
	}

	podName, containerName := e.getMasterPodInfo()
	fileUitls := operations.NewAlluxioFileUtils(podName, containerName, e.namespace, e.Log)

	ready := fileUitls.Ready()
	if !ready {
		return fmt.Errorf("the UFS is not ready")
	}

	everMounted := false
	// Iterate all the mount points, do mount if the mount point is not Fluid-native(e.g. Hostpath or PVC)
	for _, mount := range dataset.Spec.Mounts {
		mount := mount
		if common.IsFluidNativeScheme(mount.MountPoint) {
			continue
		}

		alluxioPath := utils.UFSPathBuilder{}.GenUFSPathInUnifiedNamespace(mount)

		mounted, err := fileUitls.IsMounted(alluxioPath)
		e.Log.Info("Check if the alluxio path is mounted.", "alluxioPath", alluxioPath, "mounted", mounted)
		if err != nil {
			return err
		}

		mOptions, err := e.genUFSMountOptions(mount, dataset.Spec.SharedOptions, dataset.Spec.SharedEncryptOptions, true)
		if err != nil {
			return errors.Wrapf(err, "gen ufs mount options by spec mount item failure,mount name:%s", mount.Name)
		}

		if !mounted {
			err = fileUitls.Mount(alluxioPath, mount.MountPoint, mOptions, mount.ReadOnly, mount.Shared)
			if err != nil {
				return err
			}

			everMounted = true
		}
	}

	if everMounted {
		e.updateMountTime()
	}

	return nil
}

// alluxio mount options
func (e *AlluxioEngine) genUFSMountOptions(m datav1alpha1.Mount, SharedOptions map[string]string, SharedEncryptOptions []datav1alpha1.EncryptOption,
	extractEncryptOptions bool) (map[string]string, error) {

	// initialize alluxio mount options
	mOptions := map[string]string{}
	for k, v := range SharedOptions {
		mOptions[k] = v
	}

	for key, value := range m.Options {
		mOptions[key] = value
	}

	// if encryptOptions have the same key with options, it will overwrite the corresponding value
	if extractEncryptOptions {
		// extract the encrypt options directly
		var err error
		mOptions, err = e.genEncryptOptions(SharedEncryptOptions, mOptions, m.Name)
		if err != nil {
			return mOptions, err
		}

		//gen public encryptOptions
		mOptions, err = e.genEncryptOptions(m.EncryptOptions, mOptions, m.Name)
		if err != nil {
			return mOptions, err
		}
	} else {
		// using mount file
		for _, encryptOpt := range append(SharedEncryptOptions, m.EncryptOptions...) {
			secretName := encryptOpt.ValueFrom.SecretKeyRef.Name
			secretMountPath := fmt.Sprintf("/etc/fluid/secrets/%s", secretName)
			mOptions[encryptOpt.Name] = filepath.Join(secretMountPath, encryptOpt.ValueFrom.SecretKeyRef.Key)
		}
	}
	return mOptions, nil
}

// alluxio encrypt mount options
func (e *AlluxioEngine) genEncryptOptions(EncryptOptions []datav1alpha1.EncryptOption, mOptions map[string]string, name string) (map[string]string, error) {
	for _, item := range EncryptOptions {

		if _, ok := mOptions[item.Name]; ok {
			err := fmt.Errorf("the option %s is set more than one times, please double check the dataset's option and encryptOptions", item.Name)
			return mOptions, err
		}

		securityutil.UpdateSensitiveKey(item.Name)
		sRef := item.ValueFrom.SecretKeyRef
		secret, err := kubeclient.GetSecret(e.Client, sRef.Name, e.namespace)
		if err != nil {
			e.Log.Error(err, "get secret by mount encrypt options failed", "name", item.Name)
			return mOptions, err
		}

		e.Log.Info("get value from secret", "mount name", name, "secret key", sRef.Key)

		v := secret.Data[sRef.Key]
		mOptions[item.Name] = string(v)
	}

	return mOptions, nil
}

func (e *AlluxioEngine) updateMountTime() {
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := e.getRuntime()
		if err != nil {
			return err
		}

		runtimeToUpdate := runtime.DeepCopy()
		runtimeToUpdate.Status.MountTime = &metav1.Time{Time: time.Now()}

		if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
			err = e.Client.Status().Update(context.TODO(), runtimeToUpdate)
		} else {
			e.Log.Info("Do nothing because the runtime status is not changed.")
		}

		return err
	})

	if err != nil {
		e.Log.Error(err, "UpdateMountTime", "", e.name)
	}
}
