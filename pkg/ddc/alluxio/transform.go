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
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/tieredstore"
)

func (e *AlluxioEngine) transform(runtime *datav1alpha1.AlluxioRuntime) (value *Alluxio, err error) {
	if runtime == nil {
		err = fmt.Errorf("the alluxioRuntime is null")
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

	// 7.set optimization parameters
	e.optimizeDefaultProperties(runtime, value)

	// 8.allocate port for fluid engine
	err = e.allocatePorts(value)

	// 9.set engine properties
	e.setPortProperties(runtime, value)
	return
}

// 2. Transform the common part
func (e *AlluxioEngine) transformCommonPart(runtime *datav1alpha1.AlluxioRuntime, value *Alluxio) (err error) {

	value.Image, value.ImageTag = e.parseRuntimeImage()
	// value.Image = "registry.cn-huhehaote.aliyuncs.com/alluxio/alluxio"
	if runtime.Spec.AlluxioVersion.Image != "" {
		value.Image = runtime.Spec.AlluxioVersion.Image
	}

	// value.ImageTag = "2.3.0-SNAPSHOT-238b7eb"
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

	runtimeInfo, err := e.getRuntimeInfo()
	if err != nil {
		return err
	}

	for _, level := range runtimeInfo.GetTieredstore().Levels {

		// l := 0
		// if level.MediumType == common.SSD {
		// 	l = 1
		// } else if level.MediumType == common.HDD {
		// 	l = 2
		// }

		l := tieredstore.GetTieredLevel(runtimeInfo, level.MediumType)

		levels = append(levels, Level{
			Alias:      string(level.MediumType),
			Level:      l,
			Type:       "hostPath",
			Path:       fmt.Sprintf("%s/%s/%s", level.Path, runtime.Namespace, runtime.Name),
			Mediumtype: string(level.MediumType),
			Low:        level.Low,
			High:       level.High,
			Quota:      utils.TranformQuantityToAlluxioUnit(level.Quota),
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

	e.optimizeDefaultForMaster(runtime, value)

	if len(runtime.Spec.Master.Env) > 0 {
		value.Master.Env = runtime.Spec.Master.Env
	} else {
		value.Master.Env = map[string]string{}
	}

	value.Master.Env["ALLUXIO_WORKER_TIEREDSTORE_LEVEL0_DIRS_PATH"] = value.getTiredStoreLevel0Path(e.name, e.namespace)

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
	e.optimizeDefaultForWorker(runtime, value)

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

	value.Worker.Env["ALLUXIO_WORKER_TIEREDSTORE_LEVEL0_DIRS_PATH"] = value.getTiredStoreLevel0Path(e.name, e.namespace)

	value.Worker.HostNetwork = true

	e.transformResourcesForWorker(runtime, value)

	return
}

// 8.allocate port for fluid engine
func (e *AlluxioEngine) allocatePorts(value *Alluxio) error {
	allocatedPorts, err := e.getAvaliablePort()

	if len(allocatedPorts) == PORT_NUM {
		value.Master.Ports.Rpc = allocatedPorts[0]
		value.Master.Ports.Web = allocatedPorts[1]
		value.Worker.Ports.Rpc = allocatedPorts[2]
		value.Worker.Ports.Web = allocatedPorts[3]
		value.JobMaster.Ports.Rpc = allocatedPorts[4]
		value.JobMaster.Ports.Web = allocatedPorts[5]
		value.JobWorker.Ports.Rpc = allocatedPorts[6]
		value.JobWorker.Ports.Web = allocatedPorts[7]
		value.JobWorker.Ports.Data = allocatedPorts[8]
	} else {
		value.Master.Ports.Embedded = allocatedPorts[9]
		value.JobMaster.Ports.Embedded = allocatedPorts[10]
	}

	return err
}
