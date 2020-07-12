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

	datav1alpha1 "github.com/cloudnativefluid/fluid/api/v1alpha1"
	"github.com/cloudnativefluid/fluid/pkg/common"
	"github.com/cloudnativefluid/fluid/pkg/utils"
	"github.com/cloudnativefluid/fluid/pkg/utils/tieredstore"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/resource"
)

func (e *AlluxioEngine) transform(runtime *datav1alpha1.AlluxioRuntime) (value *Alluxio, err error) {
	if runtime == nil {
		err = fmt.Errorf("The alluxioRuntime is nulll")
		return
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

	// 3.transform the fuse
	err = e.transformFuse(runtime, value)
	if err != nil {
		return
	}

	return
}

// 2. Transform the common part
func (e *AlluxioEngine) transformCommonPart(runtime *datav1alpha1.AlluxioRuntime, value *Alluxio) (err error) {

	value.Image = "alluxio/alluxio"
	if runtime.Spec.AlluxioVersion.Image != "" {
		value.Image = runtime.Spec.AlluxioVersion.Image
	}

	value.ImageTag = "2.2.1"
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

	// TODO: support nodeAffinity
	// nodeAffinity := runtime.Spec.Placement.All().NodeAffinity
	// if nodeAffinity != nil {

	// }

	if len(runtime.Spec.Properties) > 0 {
		value.Properties = runtime.Spec.Properties
	}

	if len(runtime.Spec.JvmOptions) > 0 {
		value.JvmOptions = runtime.Spec.JvmOptions
	}

	value.Fuse.ShortCircuitPolicy = "local"

	// TODO: support JVMOpitons from string to array
	// if len(runtime.Spec.JvmOptions) > 0 {
	// 	value.JvmOptions = strings.Join(runtime.Spec.JvmOptions, " ")
	// }

	// value.EnablePillars = true
	levels := []Level{}
	for _, level := range runtime.Spec.Tieredstore.Levels {

		l := 0
		if level.MediumType == common.SSD {
			l = 1
		} else if level.MediumType == common.HDD {
			l = 2
		}

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

	if len(runtime.Spec.Master.Properties) > 0 {
		value.Master.Properties = runtime.Spec.Master.Properties
	}

	value.Master.HostNetwork = true

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

	value.Worker.HostNetwork = true

	value.Worker.Resources = utils.TransformRequirementsToResources(runtime.Spec.Fuse.Resources)

	storageMap := tieredstore.GetLevelStorageMap(runtime)

	for key, requirement := range storageMap {
		if value.Worker.Resources.Limits == nil {
			value.Worker.Resources.Limits = make(common.ResourceList)
		}
		if key == common.MemoryCacheStore {
			e.Log.Info("Update the memory requirement")

			req := requirement.DeepCopy()

			if quantity, exists := runtime.Spec.Worker.Resources.Limits[corev1.ResourceMemory]; !exists || quantity.IsZero() {
				req.Add(resource.MustParse("4Gi"))
			} else {
				req.Add(quantity)
			}

			e.Log.Info("update the requiremnet", "requirement", req)

			value.Worker.Resources.Limits[corev1.ResourceMemory] = req.String()
		}
	}

	return
}

// 4. Transform the fuse
func (e *AlluxioEngine) transformFuse(runtime *datav1alpha1.AlluxioRuntime, value *Alluxio) (err error) {
	value.Fuse = Fuse{}

	value.Fuse.Image = "alluxio/alluxio-fuse"
	if runtime.Spec.Fuse.Image != "" {
		value.Fuse.Image = runtime.Spec.Fuse.Image
	}

	value.Fuse.ImageTag = "2.2.1"
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
	}

	labelName := e.getCommonLabelname()
	if len(value.Fuse.NodeSelector) == 0 {
		value.Fuse.NodeSelector = map[string]string{}
	}
	value.Fuse.NodeSelector[labelName] = "true"

	value.Fuse.HostNetwork = true
	value.Fuse.Enabled = true

	value.Fuse.Resources = utils.TransformRequirementsToResources(runtime.Spec.Fuse.Resources)

	storageMap := tieredstore.GetLevelStorageMap(runtime)

	for key, requirement := range storageMap {
		if value.Fuse.Resources.Limits == nil {
			value.Fuse.Resources.Limits = make(common.ResourceList)
		}
		if key == common.MemoryCacheStore {
			e.Log.Info("Update the memory requirement")

			req := requirement.DeepCopy()

			if quantity, exists := runtime.Spec.Fuse.Resources.Limits[corev1.ResourceMemory]; !exists || quantity.IsZero() {
				req.Add(resource.MustParse("4Gi"))
			} else {
				req.Add(quantity)
			}

			e.Log.Info("update the requiremnet", "requirement", req)

			value.Fuse.Resources.Limits[corev1.ResourceMemory] = req.String()
		}
	}

	return

}
