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
	"strings"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/tieredstore"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/resource"
)

func (e *AlluxioEngine) transform(runtime *datav1alpha1.AlluxioRuntime) (value *Alluxio, err error) {
	if runtime == nil {
		err = fmt.Errorf("The alluxioRuntime is null")
		return
	}

	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		return value, err
	}

	value = &Alluxio{}

	value.FullnameOverride = e.name

	// 1.transform the common part
	err = e.transformCommonPart(runtime, value)
	if err != nil {
		return
	}

	// 2.transform the masters
	err = e.transformMasters(runtime, value)
	if err != nil {
		return
	}

	// 3.transform the workers
	err = e.transformWorkers(runtime, value)
	if err != nil {
		return
	}

	// 4.transform the fuse
	err = e.transformFuse(runtime, dataset, value)
	if err != nil {
		return
	}

	// 5.transform the dataset if it has local path or volume
	e.transformDatasetToVolume(runtime, dataset, value)

	// 6.transform the permission
	e.transformPermission(runtime, value)

	e.optimizeDefaultProperties(runtime, value)

	return
}

// 2. Transform the common part
func (e *AlluxioEngine) transformCommonPart(runtime *datav1alpha1.AlluxioRuntime, value *Alluxio) (err error) {

	value.Image = "registry.cn-huhehaote.aliyuncs.com/alluxio/alluxio"
	if runtime.Spec.AlluxioVersion.Image != "" {
		value.Image = runtime.Spec.AlluxioVersion.Image
	}

	value.ImageTag = "2.3.0-SNAPSHOT-c8a46e3"
	if runtime.Spec.AlluxioVersion.ImageTag != "" {
		value.ImageTag = runtime.Spec.AlluxioVersion.ImageTag
	}

	value.ImagePullPolicy = "IfNotPresent"
	if runtime.Spec.AlluxioVersion.ImagePullPolicy != "" {
		value.ImagePullPolicy = runtime.Spec.AlluxioVersion.ImagePullPolicy
	}

	value.UserInfo = UserInfo{
		User:    0,
		FSGroup: 0,
		Group:   0,
	}

	// transform init users
	e.transformInitUsers(runtime, value)

	// TODO: support nodeAffinity
	// nodeAffinity := runtime.Spec.Placement.All().NodeAffinity
	// if nodeAffinity != nil {

	// }

	if len(runtime.Spec.Properties) > 0 {
		value.Properties = runtime.Spec.Properties
	} else {
		value.Properties = map[string]string{}
	}

	dataReplicas := runtime.Spec.Data.Replicas
	if dataReplicas <= 0 {
		dataReplicas = 1
	}
	// Set the max replication
	value.Properties["alluxio.user.file.replication.max"] = fmt.Sprintf("%d", dataReplicas)

	// set default storage
	value.Properties["alluxio.master.mount.table.root.ufs"] = e.getLocalStorageDirectory()

	if len(runtime.Spec.JvmOptions) > 0 {
		value.JvmOptions = runtime.Spec.JvmOptions
	}

	value.Fuse.ShortCircuitPolicy = "local"

	// TODO: support JVMOpitons from string to array
	// if len(runtime.Spec.JvmOptions) > 0 {
	// 	value.JvmOptions = strings.Join(runtime.Spec.JvmOptions, " ")
	// }

	// value.Enablefluid = true
	levels := []Level{}
	for _, level := range runtime.Spec.Tieredstore.Levels {

		// l := 0
		// if level.MediumType == common.SSD {
		// 	l = 1
		// } else if level.MediumType == common.HDD {
		// 	l = 2
		// }

		l := tieredstore.GetTieredLevel(runtime, level.MediumType)

		levels = append(levels, Level{
			Alias:      string(level.MediumType),
			Level:      l,
			Type:       "hostPath",
			Path:       level.Path,
			Mediumtype: string(level.MediumType),
			Low:        level.Low,
			High:       level.High,
			Quota:      tranformQuantityToAlluxioUnit(level.Quota),
		})
	}

	value.Tieredstore.Levels = levels

	// value.Metastore = Metastore{
	// 	VolumeType: "emptyDir",
	// }

	// quantity, err := resource.ParseQuantity("10Gi")
	// if err != nil {
	// 	return err
	// }

	value.Journal = Journal{
		VolumeType: "emptyDir",
		Size:       "30Gi",
	}

	value.ShortCircuit = ShortCircuit{
		VolumeType: "emptyDir",
		Policy:     "local",
		Enable:     true,
	}

	return
}

// 2. Transform the masters
func (e *AlluxioEngine) transformMasters(runtime *datav1alpha1.AlluxioRuntime, value *Alluxio) (err error) {

	value.Master = Master{}
	if runtime.Spec.Master.Replicas == 0 {
		value.Master.Replicas = 1
	} else {
		value.Master.Replicas = runtime.Spec.Master.Replicas
	}

	// if len(runtime.Spec.Master.JvmOptions) > 0 {
	// 	value.Master.JvmOptions = strings.Join(runtime.Spec.Master.JvmOptions, " ")
	// }
	if len(value.Master.JvmOptions) > 0 {
		value.Master.JvmOptions = runtime.Spec.Master.JvmOptions
	}

	if len(runtime.Spec.Master.Env) > 0 {
		value.Master.Env = runtime.Spec.Master.Env
	} else {
		value.Master.Env = map[string]string{}
	}

	value.Master.Env["ALLUXIO_WORKER_TIEREDSTORE_LEVEL0_DIRS_PATH"] = value.getTiredStoreLevel0Path()

	if len(runtime.Spec.Master.Properties) > 0 {
		value.Master.Properties = runtime.Spec.Master.Properties
	}

	value.Master.HostNetwork = true

	// // check the run as
	// if runtime.Spec.RunAs != nil {
	// 	value.Master.Env["ALLUXIO_USERNAME"] = alluxioUser
	// 	value.Master.Env["ALLUXIO_GROUP"] = alluxioUser
	// 	value.Master.Env["ALLUXIO_UID"] = strconv.FormatInt(*runtime.Spec.RunAs.UID, 10)
	// 	value.Master.Env["ALLUXIO_GID"] = strconv.FormatInt(*runtime.Spec.RunAs.GID, 10)
	// }

	return
}

// 3. Transform the workers
func (e *AlluxioEngine) transformWorkers(runtime *datav1alpha1.AlluxioRuntime, value *Alluxio) (err error) {
	value.Worker = Worker{}
	if len(runtime.Spec.Worker.JvmOptions) > 0 {
		value.Worker.JvmOptions = runtime.Spec.Worker.JvmOptions
	}

	// labelName := common.LabelAnnotationStorageCapacityPrefix + e.runtimeType + "-" + e.name
	labelName := e.getCommonLabelname()

	if len(value.Worker.NodeSelector) == 0 {
		value.Worker.NodeSelector = map[string]string{}
	}
	value.Worker.NodeSelector[labelName] = "true"

	if len(runtime.Spec.Worker.Properties) > 0 {
		value.Worker.Properties = runtime.Spec.Worker.Properties
	}

	if len(runtime.Spec.Worker.Env) > 0 {
		value.Worker.Env = runtime.Spec.Worker.Env
	} else {
		value.Worker.Env = map[string]string{}
	}

	// check the run as
	// if runtime.Spec.RunAs != nil {
	// 	value.Worker.Env["ALLUXIO_USERNAME"] = alluxioUser
	// 	value.Worker.Env["ALLUXIO_GROUP"] = alluxioUser
	// 	value.Worker.Env["ALLUXIO_UID"] = strconv.FormatInt(*runtime.Spec.RunAs.UID, 10)
	// 	value.Worker.Env["ALLUXIO_GID"] = strconv.FormatInt(*runtime.Spec.RunAs.GID, 10)
	// }

	value.Worker.Env["ALLUXIO_WORKER_TIEREDSTORE_LEVEL0_DIRS_PATH"] = value.getTiredStoreLevel0Path()

	value.Worker.HostNetwork = true

	value.Worker.Resources = utils.TransformRequirementsToResources(runtime.Spec.Worker.Resources)

	storageMap := tieredstore.GetLevelStorageMap(runtime)

	e.Log.Info("transformWorkers", "storageMap", storageMap)

	// TODO(iluoeli): it should be xmx + direct memory
	memLimit := resource.MustParse("20Gi")
	if quantity, exists := runtime.Spec.Worker.Resources.Limits[corev1.ResourceMemory]; exists && !quantity.IsZero() {
		memLimit = quantity
	}

	for key, requirement := range storageMap {
		if value.Worker.Resources.Limits == nil {
			value.Worker.Resources.Limits = make(common.ResourceList)
		}
		if key == common.MemoryCacheStore {
			req := requirement.DeepCopy()

			memLimit.Add(req)

			e.Log.Info("update the requirement for memory", "requirement", memLimit)

		}
		// } else if key == common.DiskCacheStore {
		// 	req := requirement.DeepCopy()

		// 	e.Log.Info("update the requiremnet for disk", "requirement", req)

		// 	value.Worker.Resources.Limits[corev1.ResourceEphemeralStorage] = req.String()
		// }
	}

	value.Worker.Resources.Limits[corev1.ResourceMemory] = memLimit.String()

	return
}

// 4. Transform the fuse
func (e *AlluxioEngine) transformFuse(runtime *datav1alpha1.AlluxioRuntime, dataset *datav1alpha1.Dataset, value *Alluxio) (err error) {
	value.Fuse = Fuse{}

	value.Fuse.Image = "registry.cn-huhehaote.aliyuncs.com/alluxio/alluxio-fuse"
	if runtime.Spec.Fuse.Image != "" {
		value.Fuse.Image = runtime.Spec.Fuse.Image
	}

	value.Fuse.ImageTag = "2.3.0-SNAPSHOT-c8a46e3"
	if runtime.Spec.Fuse.ImageTag != "" {
		value.Fuse.ImageTag = runtime.Spec.Fuse.ImageTag
	}

	value.Fuse.ImagePullPolicy = "IfNotPresent"
	if runtime.Spec.Fuse.ImagePullPolicy != "" {
		value.Fuse.ImagePullPolicy = runtime.Spec.Fuse.ImagePullPolicy
	}

	if len(runtime.Spec.Fuse.Properties) > 0 {
		value.Fuse.Properties = runtime.Spec.Fuse.Properties
	}

	// TODO: support JVMOpitons from string to array
	if len(runtime.Spec.Fuse.JvmOptions) > 0 {
		value.Fuse.JvmOptions = runtime.Spec.Fuse.JvmOptions
	}

	if len(runtime.Spec.Fuse.Env) > 0 {
		value.Fuse.Env = runtime.Spec.Fuse.Env
	} else {
		value.Fuse.Env = map[string]string{}
	}

	// if runtime.Spec.Fuse.MountPath != "" {
	// 	value.Fuse.MountPath = runtime.Spec.Fuse.MountPath
	// } else {
	// 	value.Fuse.MountPath = fmt.Sprintf("format", a)
	// }

	value.Fuse.MountPath = e.getMountPoint()
	value.Fuse.Env["MOUNT_POINT"] = value.Fuse.MountPath

	if len(runtime.Spec.Fuse.Args) > 0 {
		value.Fuse.Args = runtime.Spec.Fuse.Args
	} else {
		value.Fuse.Args = []string{"fuse", "--fuse-opts=kernel_cache"}
	}

	if dataset.Spec.Owner != nil {
		value.Fuse.Args[len(value.Fuse.Args)-1] = strings.Join([]string{value.Fuse.Args[len(value.Fuse.Args)-1], fmt.Sprintf("uid=%d,gid=%d", *dataset.Spec.Owner.UID, *dataset.Spec.Owner.GID)}, ",")
	} else {
		if len(value.Properties) == 0 {
			value.Properties = map[string]string{}
		}
		value.Properties["alluxio.fuse.user.group.translation.enabled"] = "true"
	}
	// value.Fuse.Args[-1]

	labelName := e.getCommonLabelname()
	if len(value.Fuse.NodeSelector) == 0 {
		value.Fuse.NodeSelector = map[string]string{}
	}
	value.Fuse.NodeSelector[labelName] = "true"

	value.Fuse.HostNetwork = true
	value.Fuse.Enabled = true

	value.Fuse.Resources = utils.TransformRequirementsToResources(runtime.Spec.Fuse.Resources)

	storageMap := tieredstore.GetLevelStorageMap(runtime)

	e.Log.Info("transformFuse", "storageMap", storageMap)

	// TODO(iluoeli): it should be xmx + direct memory
	memLimit := resource.MustParse("50Gi")
	if quantity, exists := runtime.Spec.Fuse.Resources.Limits[corev1.ResourceMemory]; exists && !quantity.IsZero() {
		memLimit = quantity
	}

	for key, requirement := range storageMap {
		if value.Fuse.Resources.Limits == nil {
			value.Fuse.Resources.Limits = make(common.ResourceList)
		}
		if key == common.MemoryCacheStore {
			req := requirement.DeepCopy()

			memLimit.Add(req)

			e.Log.Info("update the requiremnet for memory", "requirement", memLimit)

		}
		// } else if key == common.DiskCacheStore {
		// 	req := requirement.DeepCopy()
		// 	e.Log.Info("update the requiremnet for disk", "requirement", req)
		// 	value.Fuse.Resources.Limits[corev1.ResourceEphemeralStorage] = req.String()
		// }
	}
	if value.Fuse.Resources.Limits != nil {
		value.Fuse.Resources.Limits[corev1.ResourceMemory] = memLimit.String()
	}

	return

}
