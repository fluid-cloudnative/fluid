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

package utils

import (
	"context"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//GetDataset gets the dataset.
//It returns a pointer to the dataset if successful.
func GetDataset(client client.Client, name, namespace string) (*datav1alpha1.Dataset, error) {

	key := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	var dataset datav1alpha1.Dataset
	if err := client.Get(context.TODO(), key, &dataset); err != nil {
		return nil, err
	}
	return &dataset, nil
}

// checks the setup is done
func IsSetupDone(dataset *datav1alpha1.Dataset) (done bool) {
	index, _ := GetDatasetCondition(dataset.Status.Conditions, datav1alpha1.DatasetReady)
	if index != -1 {
		// e.Log.V(1).Info("The runtime is already setup.")
		done = true
	}

	return
}

func GetAccessModesOfDataset(client client.Client, name, namespace string) (accessModes []v1.PersistentVolumeAccessMode, err error) {
	dataset, err := GetDataset(client, name, namespace)
	if err != nil {
		return accessModes, err
	}

	accessModes = dataset.Spec.AccessModes
	if len(accessModes) == 0 {
		accessModes = []v1.PersistentVolumeAccessMode{
			v1.ReadOnlyMany,
		}
	}

	return accessModes, err

}

// IsTargetPathUnderFluidNativeMounts checks if targetPath is a subpath under some given native mount point.
// We check this for the reason that native mount points need extra metadata sync alluxioOperations.
func IsTargetPathUnderFluidNativeMounts(targetPath string, dataset datav1alpha1.Dataset) bool {
	for _, mount := range dataset.Spec.Mounts {

		mPath := UFSPathBuilder{}.GenAlluxioMountPath(mount, dataset.Spec.Mounts)

		//TODO(xuzhihao): HasPrefix is not enough.

		// only for native scheme
		if !common.IsFluidNativeScheme(mount.MountPoint) {
			continue
		}

		if IsSubPath(mPath, targetPath) {
			return true
		}
	}

	return false
}
