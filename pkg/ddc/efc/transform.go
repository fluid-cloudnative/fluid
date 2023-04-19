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

package efc

import (
	"fmt"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/transfromer"
	corev1 "k8s.io/api/core/v1"
)

func (e *EFCEngine) transform(runtime *datav1alpha1.EFCRuntime) (value *EFC, err error) {
	if runtime == nil {
		err = fmt.Errorf("the efcRuntime is null")
		return
	}

	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		return value, err
	}

	value = &EFC{
		// Set ownerReference to all EFCRuntime resources
		Owner: transfromer.GenerateOwnerReferenceFromObject(runtime),
	}

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

	err = e.transformInitFuse(runtime, value)
	if err != nil {
		return
	}

	e.transformOSAdvice(runtime, value)
	e.transformPlacementMode(dataset, value)
	e.transformTolerations(dataset, value)

	err = e.transformPodMetadata(runtime, value)
	if err != nil {
		return
	}

	return
}

func (e *EFCEngine) transformMasters(runtime *datav1alpha1.EFCRuntime,
	dataset *datav1alpha1.Dataset,
	value *EFC) (err error) {
	value.Master = Master{}

	mountInfo, err := e.getMountInfo()
	if err != nil {
		return
	}
	value.Master.MountPoint = mountInfo.MountPoint

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
	err = e.transformPortForMaster(runtime, value)
	if err != nil {
		return
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

func (e *EFCEngine) transformWorkers(runtime *datav1alpha1.EFCRuntime,
	value *EFC) (err error) {
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
	err = e.transformPortForWorker(runtime, value)
	if err != nil {
		return
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

func (e *EFCEngine) transformFuse(runtime *datav1alpha1.EFCRuntime,
	dataset *datav1alpha1.Dataset,
	value *EFC) (err error) {
	value.Fuse = Fuse{}

	mountInfo, err := e.getMountInfo()
	if err != nil {
		return
	}
	value.Fuse.MountPoint = mountInfo.MountPoint

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
	err = e.transformPortForFuse(runtime, value)
	if err != nil {
		return
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

func (e *EFCEngine) transformInitFuse(runtime *datav1alpha1.EFCRuntime,
	value *EFC) (err error) {
	value.InitFuse = InitFuse{}

	image := runtime.Spec.InitFuse.Version.Image
	tag := runtime.Spec.InitFuse.Version.ImageTag
	imagePullPolicy := runtime.Spec.InitFuse.Version.ImagePullPolicy
	value.InitFuse.Image, value.InitFuse.ImageTag, value.InitFuse.ImagePullPolicy = e.parseInitFuseImage(image, tag, imagePullPolicy)

	return nil
}

func (e *EFCEngine) transformOSAdvice(runtime *datav1alpha1.EFCRuntime,
	value *EFC) {
	value.OSAdvise.OSVersion = runtime.Spec.OSAdvise.OSVersion
	value.OSAdvise.Enabled = runtime.Spec.OSAdvise.Enabled
}

func (e *EFCEngine) transformPlacementMode(dataset *datav1alpha1.Dataset, value *EFC) {
	value.PlacementMode = string(dataset.Spec.PlacementMode)
	if len(value.PlacementMode) == 0 {
		value.PlacementMode = string(datav1alpha1.ExclusiveMode)
	}
}

func (e *EFCEngine) transformTolerations(dataset *datav1alpha1.Dataset, value *EFC) {
	if len(dataset.Spec.Tolerations) > 0 {
		// value.Tolerations = dataset.Spec.Tolerations
		value.Tolerations = []corev1.Toleration{}
		for _, toleration := range dataset.Spec.Tolerations {
			toleration.TolerationSeconds = nil
			value.Tolerations = append(value.Tolerations, toleration)
		}
	}
}

func (e *EFCEngine) transformPodMetadata(runtime *datav1alpha1.EFCRuntime, value *EFC) (err error) {
	commonLabels := utils.UnionMapsWithOverride(map[string]string{}, runtime.Spec.PodMetadata.Labels)
	value.Master.Labels = utils.UnionMapsWithOverride(commonLabels, runtime.Spec.Master.PodMetadata.Labels)
	value.Worker.Labels = utils.UnionMapsWithOverride(commonLabels, runtime.Spec.Worker.PodMetadata.Labels)
	value.Fuse.Labels = utils.UnionMapsWithOverride(commonLabels, runtime.Spec.Fuse.PodMetadata.Labels)

	commonAnnotations := utils.UnionMapsWithOverride(map[string]string{}, runtime.Spec.PodMetadata.Annotations)
	value.Master.Annotations = utils.UnionMapsWithOverride(commonAnnotations, runtime.Spec.Master.PodMetadata.Annotations)
	value.Worker.Annotations = utils.UnionMapsWithOverride(commonAnnotations, runtime.Spec.Worker.PodMetadata.Annotations)
	value.Fuse.Annotations = utils.UnionMapsWithOverride(commonAnnotations, runtime.Spec.Fuse.PodMetadata.Annotations)

	return nil
}
