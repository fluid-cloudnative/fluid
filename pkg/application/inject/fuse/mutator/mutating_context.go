/*
Copyright 2023 The Fluid Authors.

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

package mutator

import (
	"fmt"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilpointer "k8s.io/utils/pointer"
)

type MutatingPodSpecs struct {
	Volumes        []corev1.Volume
	VolumeMounts   []corev1.VolumeMount
	Containers     []corev1.Container
	InitContainers []corev1.Container
	MetaObj        metav1.ObjectMeta
}

func CollectFluidObjectSpecs(pod common.FluidObject) (*MutatingPodSpecs, error) {
	volumes, err := pod.GetVolumes()
	if err != nil {
		return nil, err
	}

	volumeMounts, err := pod.GetVolumeMounts()
	if err != nil {
		return nil, err
	}

	containers, err := pod.GetContainers()
	if err != nil {
		return nil, err
	}

	initContainers, err := pod.GetInitContainers()
	if err != nil {
		return nil, err
	}

	metaObj, err := pod.GetMetaObject()
	if err != nil {
		return nil, err
	}

	return &MutatingPodSpecs{
		Volumes:        volumes,
		VolumeMounts:   volumeMounts,
		Containers:     containers,
		InitContainers: initContainers,
		MetaObj:        metaObj,
	}, nil
}

func ApplyFluidObjectSpecs(pod common.FluidObject, specs *MutatingPodSpecs) error {
	if err := pod.SetVolumes(specs.Volumes); err != nil {
		return err
	}

	if err := pod.SetContainers(specs.Containers); err != nil {
		return err
	}

	if err := pod.SetInitContainers(specs.InitContainers); err != nil {
		return err
	}

	if err := pod.SetMetaObject(specs.MetaObj); err != nil {
		return err
	}

	return nil
}

type mutatingContext struct {
	appendedVolumeNames         map[string]string
	datasetUsedInContainers     *bool
	datasetUsedInInitContainers *bool
}

func (ctx *mutatingContext) GetAppendedVolumeNames() (nameMapping map[string]string, err error) {
	if ctx.appendedVolumeNames == nil {
		ctx.appendedVolumeNames = map[string]string{}
	}

	return ctx.appendedVolumeNames, nil
}

func (ctx *mutatingContext) SetAppendedVolumeNames(nameMapping map[string]string) {
	ctx.appendedVolumeNames = nameMapping
}

func (ctx *mutatingContext) GetDatsetUsedInContainers() (bool, error) {
	if ctx.datasetUsedInContainers == nil {
		return false, fmt.Errorf("mutatingContext.datasetUsedInContainers is not set")
	}

	return *ctx.datasetUsedInContainers, nil
}

func (ctx *mutatingContext) SetDatasetUsedInContainers(used bool) {
	ctx.datasetUsedInContainers = utilpointer.BoolPtr(used)
}

func (ctx *mutatingContext) GetDatasetUsedInInitContainers() (bool, error) {
	if ctx.datasetUsedInInitContainers == nil {
		return false, fmt.Errorf("mutatingContext.datasetUsedInInitContainers is not set")
	}

	return *ctx.datasetUsedInInitContainers, nil
}

func (ctx *mutatingContext) SetDatasetUsedInInitContainers(used bool) {
	ctx.datasetUsedInInitContainers = utilpointer.Bool(used)
}
