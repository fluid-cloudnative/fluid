/*
Copyright 2024 The Fluid Authors.
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

package vineyard

import (
	"fmt"
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/transfromer"
	"k8s.io/apimachinery/pkg/api/resource"
)

func (e *VineyardEngine) transform(runtime *datav1alpha1.VineyardRuntime) (value *Vineyard, err error) {
	if runtime == nil {
		err = fmt.Errorf("the vineyardRuntime is null")
		return
	}
	defer utils.TimeTrack(time.Now(), "VineyardRuntime.Transform", "name", runtime.Name)

	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		return value, err
	}

	value = &Vineyard{
		Owner: transfromer.GenerateOwnerReferenceFromObject(runtime),
	}
	value.FullnameOverride = e.name

	value.TieredStore = e.transformTieredStore(runtime)
	err = e.transformMasters(runtime, dataset, value)
	if err != nil {
		return
	}

	err = e.transformWorkers(runtime, value)
	if err != nil {
		return
	}

	e.transformFuse(runtime, value)
	return value, nil
}

func (e *VineyardEngine) transformMasters(runtime *datav1alpha1.VineyardRuntime,
	dataset *datav1alpha1.Dataset,
	value *Vineyard,
) (err error) {
	value.Master = Master{}
	if runtime.Spec.Master.Replicas == 0 {
		value.Master.Replicas = 1
	} else {
		value.Master.Replicas = runtime.Spec.Master.Replicas
	}
	image := runtime.Spec.Master.Image
	imageTag := runtime.Spec.Master.ImageTag
	imagePullPolicy := runtime.Spec.Master.ImagePullPolicy

	value.Master.Image, value.Master.ImageTag, value.Master.ImagePullPolicy = e.parseMasterImage(image, imageTag, imagePullPolicy)

	if len(runtime.Spec.Master.Env) > 0 {
		value.Master.Env = runtime.Spec.Master.Env
	} else {
		value.Master.Env = map[string]string{}
	}
	options := e.transformMasterOptions(runtime)
	if len(options) != 0 {
		value.Master.Options = options
	}

	nodeSelector := e.transformMasterSelector(runtime)
	if len(nodeSelector) != 0 {
		value.Master.NodeSelector = nodeSelector
	}

	e.transformResourcesForMaster(runtime, value)

	ports := e.transformMasterPorts(runtime)
	if len(ports) != 0 {
		value.Master.Ports = ports
	}

	err = e.transformMasterVolumes(runtime, value)
	if err != nil {
		e.Log.Error(err, "failed to transform volumes for master")
	}

	return
}

func (e *VineyardEngine) transformWorkers(runtime *datav1alpha1.VineyardRuntime, value *Vineyard) (err error) {
	value.Worker = Worker{}
	if runtime.Spec.Worker.Replicas == 0 {
		value.Worker.Replicas = 1
	} else {
		value.Worker.Replicas = runtime.Spec.Worker.Replicas
	}
	image := runtime.Spec.Worker.Image
	imageTag := runtime.Spec.Worker.ImageTag
	imagePullPolicy := runtime.Spec.Worker.ImagePullPolicy

	value.Worker.Image, value.Worker.ImageTag, value.Worker.ImagePullPolicy = e.parseWorkerImage(image, imageTag, imagePullPolicy)

	if len(runtime.Spec.Worker.Env) > 0 {
		value.Worker.Env = runtime.Spec.Worker.Env
	} else {
		value.Worker.Env = map[string]string{}
	}

	if len(runtime.Spec.Worker.NodeSelector) > 0 {
		value.Worker.NodeSelector = runtime.Spec.Worker.NodeSelector
	} else {
		value.Worker.NodeSelector = map[string]string{}
	}

	if err := e.transformResourcesForWorker(runtime, value); err != nil {
		return err
	}

	ports := e.transformWorkerPorts(runtime)
	if len(ports) != 0 {
		value.Worker.Ports = ports
	}

	err = e.transformWorkerVolumes(runtime, value)
	if err != nil {
		e.Log.Error(err, "failed to transform volumes for worker")
	}

	return
}

func (e *VineyardEngine) transformFuse(runtime *datav1alpha1.VineyardRuntime, value *Vineyard) {
	value.Fuse = Fuse{}
	image := runtime.Spec.Fuse.Image
	imageTag := runtime.Spec.Fuse.ImageTag
	imagePullPolicy := runtime.Spec.Fuse.ImagePullPolicy
	value.Fuse.Image, value.Fuse.ImageTag, value.Fuse.ImagePullPolicy = e.parseFuseImage(image, imageTag, imagePullPolicy)

	value.Fuse.CleanPolicy = runtime.Spec.Fuse.CleanPolicy

	value.Fuse.NodeSelector = e.transformFuseNodeSelector(runtime)

	value.Fuse.TargetPath = e.getMountPoint()
	e.transformResourcesForFuse(runtime, value)
}

func (e *VineyardEngine) transformMasterSelector(runtime *datav1alpha1.VineyardRuntime) map[string]string {
	properties := map[string]string{}
	if runtime.Spec.Master.NodeSelector != nil {
		properties = runtime.Spec.Master.NodeSelector
	}
	return properties
}

func (e *VineyardEngine) transformMasterPorts(runtime *datav1alpha1.VineyardRuntime) map[string]int {
	ports := map[string]int{
		MasterClientName: MasterClientPort,
		MasterPeerName:   MasterPeerPort,
	}
	if len(runtime.Spec.Master.Ports) > 0 {
		for key, value := range runtime.Spec.Master.Ports {
			ports[key] = value
		}
	}
	return ports
}

func (e *VineyardEngine) transformMasterOptions(runtime *datav1alpha1.VineyardRuntime) map[string]string {
	options := map[string]string{
		WorkerReserveMemory: DefaultWorkerReserveMemoryValue,
		WorkerEtcdPrefix:    DefaultWorkerEtcdPrefixValue,
	}
	if len(runtime.Spec.Master.Options) > 0 {
		for key, value := range runtime.Spec.Master.Options {
			options[key] = value
		}
	}
	return options
}

func (e *VineyardEngine) transformWorkerOptions(runtime *datav1alpha1.VineyardRuntime) map[string]string {
	options := map[string]string{}
	if len(runtime.Spec.Worker.Options) > 0 {
		options = runtime.Spec.Worker.Options
	}

	return options
}

func (e *VineyardEngine) transformWorkerPorts(runtime *datav1alpha1.VineyardRuntime) map[string]int {
	ports := map[string]int{
		WorkerRPCName:      WorkerRPCPort,
		WorkerExporterName: WorkerExporterPort,
	}
	if len(runtime.Spec.Worker.Ports) > 0 {
		for key, value := range runtime.Spec.Worker.Ports {
			ports[key] = value
		}
	}
	return ports
}

func (e *VineyardEngine) transformFuseNodeSelector(runtime *datav1alpha1.VineyardRuntime) map[string]string {
	nodeSelector := map[string]string{}
	nodeSelector[e.getFuseLabelName()] = "true"
	return nodeSelector
}

func (e *VineyardEngine) transformTieredStore(runtime *datav1alpha1.VineyardRuntime) TieredStore {
	quota := resource.MustParse("4Gi")
	tieredStore := TieredStore{
		Levels: []Level{
			{
				MediumType: "MEM",
				Level:      0,
				Quota:      &quota,
			},
		},
	}
	if len(runtime.Spec.TieredStore.Levels) != 0 {
		tieredStore = TieredStore{}
		for _, level := range runtime.Spec.TieredStore.Levels {
			if level.MediumType == "MEM" {
				tieredStore.Levels = append(tieredStore.Levels, Level{
					MediumType: level.MediumType,
					Quota:      level.Quota,
				})
			} else {
				tieredStore.Levels = append(tieredStore.Levels, Level{
					Level:        1,
					MediumType:   level.MediumType,
					VolumeType:   level.VolumeType,
					VolumeSource: level.VolumeSource,
					Path:         level.Path,
					Quota:        level.Quota,
					QuotaList:    level.QuotaList,
					High:         level.High,
					Low:          level.Low,
				})
			}
		}
	}
	return tieredStore
}
