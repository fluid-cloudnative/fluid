/*
  Copyright 2022 The Fluid Authors.

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

package eac

import (
	"errors"
	"fmt"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"strings"
)

func (e *EACEngine) transform(runtime *datav1alpha1.EACRuntime) (value *EAC, err error) {
	if runtime == nil {
		err = fmt.Errorf("the eacRuntime is null")
		return
	}

	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		return value, err
	}

	value = &EAC{}

	value.FullnameOverride = e.name

	err = e.transformMasters(runtime, dataset, value)
	if err != nil {
		return
	}

	err = e.transformWorkers(runtime, value)
	if err != nil {
		return
	}

	err = e.transformFuse(runtime, dataset, value)
	if err != nil {
		return
	}

	err = e.transformInitAliFuse(runtime, value)
	if err != nil {
		return
	}

	e.transformPlacementMode(dataset, value)
	e.transformTolerations(dataset, value)
	return
}

func (e *EACEngine) transformMasters(runtime *datav1alpha1.EACRuntime,
	dataset *datav1alpha1.Dataset,
	value *EAC) (err error) {
	value.Master = Master{}

	err = e.transformMountPoint(&value.Master.MountPoint, dataset)
	if err != nil {
		return
	}

	value.Master.Replicas = runtime.MasterReplicas()
	value.Master.Enabled = runtime.MasterEnabled()
	value.Master.HostNetwork = datav1alpha1.IsHostNetwork(runtime.Spec.Master.NetworkMode)

	// image
	image := runtime.Spec.Master.Version.Image
	tag := runtime.Spec.Master.Version.ImageTag
	imagePullPolicy := runtime.Spec.Master.Version.ImagePullPolicy
	value.Master.Image, value.Master.ImageTag, value.Master.ImagePullPolicy = e.parseMasterImage(image, tag, imagePullPolicy)

	// node selector
	value.Master.NodeSelector = map[string]string{}
	if len(runtime.Spec.Master.NodeSelector) > 0 {
		value.Master.NodeSelector = runtime.Spec.Master.NodeSelector
	}

	// tiered store
	err = e.transformMasterTieredStore(runtime, value)
	if err != nil {
		return
	}

	// ports
	if datav1alpha1.IsHostNetwork(runtime.Spec.Master.NetworkMode) {
		e.Log.Info("allocateMasterPorts for hostnetwork mode")
		err = e.allocateMasterPorts(value)
		if err != nil {
			return
		}
	} else {
		e.Log.Info("skip allocateMasterPorts for container network mode")
		e.generateMasterStaticPorts(value)
	}

	// resources
	err = e.transformResourcesForMaster(runtime, value)
	if err != nil {
		return
	}

	// options
	err = e.transformMasterOptions(runtime, value)
	if err != nil {
		return
	}

	return nil
}

func (e *EACEngine) transformWorkers(runtime *datav1alpha1.EACRuntime,
	value *EAC) (err error) {
	value.Worker = Worker{}

	value.Worker.Enabled = runtime.Enabled()
	value.Worker.HostNetwork = datav1alpha1.IsHostNetwork(runtime.Spec.Worker.NetworkMode)

	// image
	image := runtime.Spec.Worker.Version.Image
	tag := runtime.Spec.Worker.Version.ImageTag
	imagePullPolicy := runtime.Spec.Worker.Version.ImagePullPolicy
	value.Worker.Image, value.Worker.ImageTag, value.Worker.ImagePullPolicy = e.parseWorkerImage(image, tag, imagePullPolicy)

	// node selector
	value.Worker.NodeSelector = map[string]string{}
	if len(runtime.Spec.Worker.NodeSelector) > 0 {
		value.Worker.NodeSelector = runtime.Spec.Worker.NodeSelector
	}

	// tiered store
	err = e.transformWorkerTieredStore(runtime, value)
	if err != nil {
		return
	}

	// ports
	if datav1alpha1.IsHostNetwork(runtime.Spec.Worker.NetworkMode) {
		e.Log.Info("allocateWorkerPorts for hostnetwork mode")
		err = e.allocateWorkerPorts(value)
		if err != nil {
			return
		}
	} else {
		e.Log.Info("skip allocateWorkerPorts for container network mode")
		e.generateWorkerStaticPorts(value)
	}

	// resources
	err = e.transformResourcesForWorker(runtime, value)
	if err != nil {
		return
	}

	// options
	err = e.transformWorkerOptions(runtime, value)
	if err != nil {
		return
	}

	return nil
}

func (e *EACEngine) transformFuse(runtime *datav1alpha1.EACRuntime,
	dataset *datav1alpha1.Dataset,
	value *EAC) (err error) {
	value.Fuse = Fuse{}

	err = e.transformMountPoint(&value.Fuse.MountPoint, dataset)
	if err != nil {
		return
	}

	value.Fuse.HostMountPath = e.getHostMountPath()
	value.Fuse.HostNetwork = datav1alpha1.IsHostNetwork(runtime.Spec.Fuse.NetworkMode)
	value.Fuse.CriticalPod = common.CriticalFusePodEnabled()

	// image
	image := runtime.Spec.Fuse.Version.Image
	tag := runtime.Spec.Fuse.Version.ImageTag
	imagePullPolicy := runtime.Spec.Fuse.Version.ImagePullPolicy
	value.Fuse.Image, value.Fuse.ImageTag, value.Fuse.ImagePullPolicy = e.parseFuseImage(image, tag, imagePullPolicy)

	// node selector
	value.Fuse.NodeSelector = map[string]string{}
	if len(runtime.Spec.Fuse.NodeSelector) > 0 {
		value.Fuse.NodeSelector = runtime.Spec.Fuse.NodeSelector
	}
	// The label will be added by CSI Plugin when any workload pod is scheduled on the node.
	value.Fuse.NodeSelector[e.getFuseLabelName()] = "true"

	// tiered store
	err = e.transformFuseTieredStore(runtime, value)
	if err != nil {
		return
	}

	// ports
	if datav1alpha1.IsHostNetwork(runtime.Spec.Fuse.NetworkMode) {
		e.Log.Info("allocateFusePorts for hostnetwork mode")
		err = e.allocateFusePorts(value)
		if err != nil {
			return
		}
	} else {
		e.Log.Info("skip allocateFusePorts for container network mode")
		e.generateFuseStaticPorts(value)
	}

	// resources
	err = e.transformResourcesForFuse(runtime, value)
	if err != nil {
		return
	}

	// options
	err = e.transformFuseOptions(runtime, value)
	if err != nil {
		return
	}

	return nil
}

func (e *EACEngine) transformInitAliFuse(runtime *datav1alpha1.EACRuntime,
	value *EAC) (err error) {
	value.InitFuse = InitFuse{}

	image := runtime.Spec.InitFuse.Version.Image
	tag := runtime.Spec.InitFuse.Version.ImageTag
	imagePullPolicy := runtime.Spec.InitFuse.Version.ImagePullPolicy
	value.InitFuse.Image, value.InitFuse.ImageTag, value.InitFuse.ImagePullPolicy = e.parseInitFuseImage(image, tag, imagePullPolicy)

	return nil
}

func (e *EACEngine) transformPlacementMode(dataset *datav1alpha1.Dataset, value *EAC) {
	value.PlacementMode = string(dataset.Spec.PlacementMode)
	if len(value.PlacementMode) == 0 {
		value.PlacementMode = string(datav1alpha1.ExclusiveMode)
	}
}

func (e *EACEngine) transformMountPoint(mountpoint *string, dataset *datav1alpha1.Dataset) error {
	var (
		eacPrefix = "eac://"
	)
	if len(dataset.Spec.Mounts) == 0 {
		return errors.New("empty mount point")
	}
	mount := dataset.Spec.Mounts[0]
	if !strings.HasSuffix(mount.MountPoint, "/") {
		mount.MountPoint = mount.MountPoint + "/"
	}
	if !strings.HasPrefix(mount.MountPoint, eacPrefix) {
		return errors.New("invalid mount point prefix, must be eac://")
	}
	*mountpoint = strings.TrimPrefix(mount.MountPoint, eacPrefix)
	return nil
}

func (e *EACEngine) transformTolerations(dataset *datav1alpha1.Dataset, value *EAC) {
	if len(dataset.Spec.Tolerations) > 0 {
		// value.Tolerations = dataset.Spec.Tolerations
		value.Tolerations = []corev1.Toleration{}
		for _, toleration := range dataset.Spec.Tolerations {
			toleration.TolerationSeconds = nil
			value.Tolerations = append(value.Tolerations, toleration)
		}
	}
}
