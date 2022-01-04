/*

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

package goosefs

import (
	"context"
	"fmt"
	"reflect"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/goosefs/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/pkg/errors"
)

func (e *GooseFSEngine) usedStorageBytesInternal() (value int64, err error) {
	return
}

func (e *GooseFSEngine) freeStorageBytesInternal() (value int64, err error) {
	return
}

func (e *GooseFSEngine) totalStorageBytesInternal() (total int64, err error) {
	podName, containerName := e.getMasterPodInfo()

	fileUitls := operations.NewGooseFSFileUtils(podName, containerName, e.namespace, e.Log)
	_, _, total, err = fileUitls.Count("/")
	if err != nil {
		return
	}

	return
}

func (e *GooseFSEngine) totalFileNumsInternal() (fileCount int64, err error) {
	podName, containerName := e.getMasterPodInfo()

	fileUitls := operations.NewGooseFSFileUtils(podName, containerName, e.namespace, e.Log)
	fileCount, err = fileUitls.GetFileCount()
	if err != nil {
		return
	}

	return
}

// shouldMountUFS checks if there's any UFS that need to be mounted
func (e *GooseFSEngine) shouldMountUFS() (should bool, err error) {
	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	e.Log.Info("get dataset info", "dataset", dataset)
	if err != nil {
		return should, err
	}

	podName, containerName := e.getMasterPodInfo()
	fileUtils := operations.NewGooseFSFileUtils(podName, containerName, e.namespace, e.Log)

	ready := fileUtils.Ready()
	if !ready {
		should = false
		err = fmt.Errorf("the UFS is not ready")
		return should, err
	}

	// Check if any of the Mounts has not been mounted in GooseFS
	for _, mount := range dataset.Spec.Mounts {
		if common.IsFluidNativeScheme(mount.MountPoint) {
			// No need for a mount point with Fluid native scheme('local://' and 'pvc://') to be mounted
			continue
		}
		goosefsPath := utils.UFSPathBuilder{}.GenAlluxioMountPath(mount, dataset.Spec.Mounts)
		mounted, err := fileUtils.IsMounted(goosefsPath)
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

// getMounts get slice of mounted paths and expected mount paths
func (e *GooseFSEngine) getMounts() (resultInCtx []string, resultHaveMounted []string, err error) {
	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	e.Log.V(1).Info("get dataset info", "dataset", dataset)
	if err != nil {
		return resultInCtx, resultHaveMounted, err
	}

	podName, containerName := e.getMasterPodInfo()
	fileUitls := operations.NewGooseFSFileUtils(podName, containerName, e.namespace, e.Log)

	ready := fileUitls.Ready()
	if !ready {
		err = fmt.Errorf("the UFS is not ready")
		return resultInCtx, resultHaveMounted, err
	}

	// Check if any of the Mounts has not been mounted in GooseFS
	for _, mount := range dataset.Spec.Mounts {
		if common.IsFluidNativeScheme(mount.MountPoint) {
			// No need for a mount point with Fluid native scheme('local://' and 'pvc://') to be mounted
			continue
		}
		goosefsPathInCtx := utils.UFSPathBuilder{}.GenAlluxioMountPath(mount, dataset.Spec.Mounts)
		resultInCtx = append(resultInCtx, goosefsPathInCtx)
	}

	// get the mount points have been mountted
	for _, mount := range dataset.Status.Mounts {
		if common.IsFluidNativeScheme(mount.MountPoint) {
			// No need for a mount point with Fluid native scheme('local://' and 'pvc://') to be mounted
			continue
		}
		goosefsPathHaveMountted := utils.UFSPathBuilder{}.GenAlluxioMountPath(mount, dataset.Status.Mounts)
		resultHaveMounted = append(resultHaveMounted, goosefsPathHaveMountted)
	}

	return resultInCtx, resultHaveMounted, err

}

// calculateMountPointsChanges will compare diff of spec mount and status mount,
// to find need to be added mount point and removed mount point
func (e *GooseFSEngine) calculateMountPointsChanges(mountsHaveMountted []string, mountsInContext []string) ([]string, []string) {
	removed := []string{}
	added := []string{}

	for _, v := range mountsHaveMountted {
		if !ContainsString(mountsInContext, v) {
			removed = append(removed, v)
		}
	}

	for _, v := range mountsInContext {
		if !ContainsString(mountsHaveMountted, v) {
			added = append(added, v)
		}
	}

	return added, removed
}

// ContainsString returns true if a string is present in a iteratee.
func ContainsString(s []string, v string) bool {
	for _, vv := range s {
		if vv == v {
			return true
		}
	}
	return false
}

// processUpdatingUFS will mount needed mountpoint to ufs
func (e *GooseFSEngine) processUpdatingUFS(ufsToUpdate *utils.UFSToUpdate) (err error) {
	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		return err
	}

	podName, containerName := e.getMasterPodInfo()
	fileUtils := operations.NewGooseFSFileUtils(podName, containerName, e.namespace, e.Log)

	ready := fileUtils.Ready()
	if !ready {
		return fmt.Errorf("the UFS is not ready, namespace:%s,name:%s", e.namespace, e.name)
	}

	// Iterate all the mount points, do mount if the mount point is in added array
	// TODO: not allow to edit FluidNativeScheme MountPoint
	for _, mount := range dataset.Spec.Mounts {
		if common.IsFluidNativeScheme(mount.MountPoint) {
			continue
		}

		goosefsPath := utils.UFSPathBuilder{}.GenAlluxioMountPath(mount, dataset.Spec.Mounts)
		if len(ufsToUpdate.ToAdd()) > 0 && utils.ContainsString(ufsToUpdate.ToAdd(), goosefsPath) {
			mountOptions := map[string]string{}
			for key, value := range mount.Options {
				mountOptions[key] = value
			}

			// Configure mountOptions using encryptOptions
			// If encryptOptions have the same key with options, it will overwrite the corresponding value
			for _, encryptOption := range mount.EncryptOptions {
				key := encryptOption.Name
				secretKeyRef := encryptOption.ValueFrom.SecretKeyRef

				secret, err := kubeclient.GetSecret(e.Client, secretKeyRef.Name, e.namespace)
				if err != nil {
					e.Log.Info("can't get the secret",
						"namespace", e.namespace,
						"name", e.name,
						"secretName", secretKeyRef.Name)
					return err
				}

				value := secret.Data[secretKeyRef.Key]
				e.Log.Info("get value from secret",
					"namespace", e.namespace,
					"name", e.name,
					"secretName", secretKeyRef.Name)

				mountOptions[key] = string(value)
			}
			err = fileUtils.Mount(goosefsPath, mount.MountPoint, mountOptions, mount.ReadOnly, mount.Shared)
			if err != nil {
				return err
			}
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
	datasetToUpdate.Status.UfsTotal = METADATA_SYNC_NOT_DONE_MSG
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

	return nil
}

// mountUFS() mount all UFSs to GooseFS according to mount points in `dataset.Spec`. If a mount point is Fluid-native, mountUFS() will skip it.
func (e *GooseFSEngine) mountUFS() (err error) {
	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		return err
	}

	podName, containerName := e.getMasterPodInfo()
	fileUitls := operations.NewGooseFSFileUtils(podName, containerName, e.namespace, e.Log)

	ready := fileUitls.Ready()
	if !ready {
		return fmt.Errorf("the UFS is not ready")
	}

	// Iterate all the mount points, do mount if the mount point is not Fluid-native(e.g. Hostpath or PVC)
	for _, mount := range dataset.Spec.Mounts {
		mount := mount
		if common.IsFluidNativeScheme(mount.MountPoint) {
			continue
		}

		goosefsPath := utils.UFSPathBuilder{}.GenAlluxioMountPath(mount, dataset.Spec.Mounts)
		mounted, err := fileUitls.IsMounted(goosefsPath)
		e.Log.Info("Check if the goosefs path is mounted.", "goosefsPath", goosefsPath, "mounted", mounted)
		if err != nil {
			return err
		}

		mOptions, err := e.genUFSMountOptions(mount)
		if err != nil {
			return errors.Wrapf(err, "gen ufs mount options by spec mount item failure,mount name:%s", mount.Name)
		}
		if !mounted {
			err = fileUitls.Mount(goosefsPath, mount.MountPoint, mOptions, mount.ReadOnly, mount.Shared)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

// goosefs mount options
func (e *GooseFSEngine) genUFSMountOptions(m datav1alpha1.Mount) (map[string]string, error) {

	// initialize goosefs mount options
	mOptions := map[string]string{}
	if len(m.Options) > 0 {
		mOptions = m.Options
	}

	// if encryptOptions have the same key with options
	// it will overwrite the corresponding value
	for _, item := range m.EncryptOptions {

		sRef := item.ValueFrom.SecretKeyRef
		secret, err := kubeclient.GetSecret(e.Client, sRef.Name, e.namespace)
		if err != nil {
			e.Log.Error(err, "get secret by mount encrypt options failed", "name", item.Name)
			return mOptions, err
		}

		e.Log.Info("get value from secret", "mount name", m.Name, "secret key", sRef.Key)

		v := secret.Data[sRef.Key]
		mOptions[item.Name] = string(v)
	}

	return mOptions, nil
}
