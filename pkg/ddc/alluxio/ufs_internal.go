/*
Copyright 2023 The Fluid Author.

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
	"reflect"
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
		alluxioPath := utils.UFSPathBuilder{}.GenAlluxioMountPath(mount, dataset.Spec.Mounts)
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
		alluxioPath := utils.UFSPathBuilder{}.GenAlluxioMountPath(mount, dataset.Spec.Mounts)
		alluxioPaths = append(alluxioPaths, alluxioPath)
	}

	// For fluid native schema, skip mount check to avoid unnecessary system call
	if len(alluxioPaths) == 0 {
		return
	}

	return fileUtils.FindUnmountedAlluxioPaths(alluxioPaths)
}

func (e *AlluxioEngine) processUpdatingUFS(ufsToUpdate *utils.UFSToUpdate) (err error) {
	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		return err
	}

	podName, containerName := e.getMasterPodInfo()
	fileUtils := operations.NewAlluxioFileUtils(podName, containerName, e.namespace, e.Log)

	ready := fileUtils.Ready()
	if !ready {
		return fmt.Errorf("the UFS is not ready, namespace:%s,name:%s", e.namespace, e.name)
	}

	everMounted := false
	// Iterate all the mount points, do mount if the mount point is in added array
	// TODO: not allow to edit FluidNativeScheme MountPoint
	for _, mount := range dataset.Spec.Mounts {
		if common.IsFluidNativeScheme(mount.MountPoint) {
			continue
		}

		alluxioPath := utils.UFSPathBuilder{}.GenAlluxioMountPath(mount, dataset.Spec.Mounts)
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
				return err
			}

			mountOptions, err = e.genEncryptOptions(mount.EncryptOptions, mountOptions, mount.Name)
			if err != nil {
				return err
			}

			err = fileUtils.Mount(alluxioPath, mount.MountPoint, mountOptions, mount.ReadOnly, mount.Shared)
			if err != nil {
				return err
			}

			everMounted = true
		}
	}

	// unmount the mount point in the removed array
	if len(ufsToUpdate.ToRemove()) > 0 {
		for _, mountRemove := range ufsToUpdate.ToRemove() {
			err = fileUtils.UnMount(mountRemove)
			if err != nil {
				return err
			}
		}
	}
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
		return nil
	}

	if everMounted {
		e.updateMountTime()
	}

	return nil
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

		alluxioPath := utils.UFSPathBuilder{}.GenAlluxioMountPath(mount, dataset.Spec.Mounts)

		mounted, err := fileUitls.IsMounted(alluxioPath)
		e.Log.Info("Check if the alluxio path is mounted.", "alluxioPath", alluxioPath, "mounted", mounted)
		if err != nil {
			return err
		}

		mOptions, err := e.genUFSMountOptions(mount, dataset.Spec.SharedOptions, dataset.Spec.SharedEncryptOptions)
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
func (e *AlluxioEngine) genUFSMountOptions(m datav1alpha1.Mount, SharedOptions map[string]string, SharedEncryptOptions []datav1alpha1.EncryptOption) (map[string]string, error) {

	// initialize alluxio mount options
	mOptions := map[string]string{}
	for k, v := range SharedOptions {
		mOptions[k] = v
	}

	for key, value := range m.Options {
		mOptions[key] = value
	}

	// if encryptOptions have the same key with options
	// it will overwrite the corresponding value
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
