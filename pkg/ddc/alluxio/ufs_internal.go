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

package alluxio

import (
	"fmt"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/alluxio/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/pkg/errors"
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
	e.Log.Info("get dataset info", "dataset", dataset)
	if err != nil {
		return should, err
	}

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

func (e *AlluxioEngine) getMounts() (resultInCtx []string, resultHaveMounted []string, err error) {
	fmt.Println("get dataset info", "Client", e.Client)
	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	fmt.Println("get dataset info", "dataset", dataset)
	e.Log.V(1).Info("get dataset info", "dataset", dataset)
	if err != nil {
		return resultInCtx, resultHaveMounted, err
	}

	podName, containerName := e.getMasterPodInfo()
	fileUitls := operations.NewAlluxioFileUtils(podName, containerName, e.namespace, e.Log)

	ready := fileUitls.Ready()
	if !ready {
		err = fmt.Errorf("the UFS is not ready")
		return resultInCtx, resultHaveMounted, err
	}

	// Check if any of the Mounts has not been mounted in Alluxio
	for _, mount := range dataset.Spec.Mounts {
		if common.IsFluidNativeScheme(mount.MountPoint) {
			// No need for a mount point with Fluid native scheme('local://' and 'pvc://') to be mounted
			continue
		}
		alluxioPathInCtx := utils.UFSPathBuilder{}.GenAlluxioMountPath(mount, dataset.Spec.Mounts)
		resultInCtx = append(resultInCtx, alluxioPathInCtx)
	}

	// get the mount points have been mountted
	for _, mount := range dataset.Status.Mounts {
		if common.IsFluidNativeScheme(mount.MountPoint) {
			// No need for a mount point with Fluid native scheme('local://' and 'pvc://') to be mounted
			continue
		}
		alluxioPathHaveMountted := utils.UFSPathBuilder{}.GenAlluxioMountPath(mount, dataset.Status.Mounts)
		resultHaveMounted = append(resultHaveMounted, alluxioPathHaveMountted)
	}

	return resultInCtx, resultHaveMounted, err

}

func (e *AlluxioEngine) calculateMountPointsChanges(mountsHaveMountted []string, mountsInContext []string) ([]string, []string) {
	msrc := make(map[string]byte) //build index by source(exists)  getMountted
	mall := make(map[string]byte) //build index by source and target(ctx) mounts_context
	var set []string              //the intersection
	//1.build map by source
	for _, v := range mountsHaveMountted {
		msrc[v] = 0
		mall[v] = 0
	}
	//2.if length does not changed, then intersected, mall will be the union (contain all elements) in the end
	for _, v := range mountsInContext { //mounts_context  alluxioPath from dataset
		l := len(mall)
		mall[v] = 1
		if l != len(mall) { //add new
			l = len(mall)
		} else { // intersected
			set = append(set, v)
		}
	}
	//3.loop the intersection，find it in the union, if found, then delete it from the union, mall will be the complement(union - intersection)
	for _, v := range set {
		delete(mall, v)
	}
	//4.find all the element in mall(the complement) in the source, if found, then add it to delete array, else, add it to add array
	var added, removed []string
	for v := range mall {
		_, exist := msrc[v]
		if exist {
			removed = append(removed, v)
		} else {
			added = append(added, v)
		}
	}

	return added, removed
}

func (e *AlluxioEngine) processUFS(added []string, removed []string) (err error) {
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

	// Iterate all the mount points, do mount if the mount point is in added array
	for _, mount := range dataset.Spec.Mounts {
		if common.IsFluidNativeScheme(mount.MountPoint) {
			continue
		}

		alluxioPath := utils.UFSPathBuilder{}.GenAlluxioMountPath(mount, dataset.Spec.Mounts)
		if len(added) > 0 && utils.ContainsString(added, alluxioPath) {
			mountOptions := map[string]string{}
			for key, value := range mount.Options {
				mountOptions[key] = value
			}

			// Configure mountOptions using encryptOptions
			// If encryptOptions have the same key with options, it will overwrite the corresponding value
			for _, encryptOption := range mount.EncryptOptions {
				key := encryptOption.Name
				secretKeyRef := encryptOption.ValueFrom.SecretKeyRef

				secret, err := utils.GetSecret(e.Client, secretKeyRef.Name, e.namespace)
				if err != nil {
					e.Log.Info("can't get the secret")
					return err
				}

				value := secret.Data[secretKeyRef.Key]
				e.Log.Info("get value from secret")

				mountOptions[key] = string(value)
			}
			err = fileUitls.Mount(alluxioPath, mount.MountPoint, mountOptions, mount.ReadOnly, mount.Shared)
			if err != nil {
				return err
			}
		}
	}

	// unmount the mount point in the removed array
	if len(removed) > 0 {
		for _, mount_remove := range removed {
			fileUitls.UnMount(mount_remove)
		}
	}

	err = e.SyncMetadata()
	if err != nil {
		// just report this error and ignore it because SyncMetadata isn't on the critical path of Setup
		e.Log.Error(err, "SyncMetadata")
		return nil
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

		mOptions, err := e.genUFSMountOptions(mount)
		if err != nil {
			return errors.Wrapf(err, "gen ufs mount options by spec mount item failure,mount name:%s", mount.Name)
		}

		if !mounted {
			err = fileUitls.Mount(alluxioPath, mount.MountPoint, mOptions, mount.ReadOnly, mount.Shared)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

// alluxio mount options
func (e *AlluxioEngine) genUFSMountOptions(m datav1alpha1.Mount) (map[string]string, error) {

	// initialize alluxio mount options
	mOptions := map[string]string{}
	if len(m.Options) > 0 {
		mOptions = m.Options
	}

	// if encryptOptions have the same key with options
	// it will overwrite the corresponding value
	for _, item := range m.EncryptOptions {

		sRef := item.ValueFrom.SecretKeyRef
		secret, err := utils.GetSecret(e.Client, sRef.Name, e.namespace)
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
