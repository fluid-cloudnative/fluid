/*
Copyright 2021 The Fluid Authors.

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

package base

import (
	"context"
	"fmt"

	"github.com/fluid-cloudnative/fluid/pkg/scripts/poststart"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"k8s.io/apimachinery/pkg/types"
)

// GetTemplateToInjectForFuse gets template for fuse injection
func (info *RuntimeInfo) GetTemplateToInjectForFuse(pvcName string) (template *common.FuseInjectionTemplate, err error) {
	// TODO: create fuse container
	ds, err := info.getFuseDaemonset()
	if err != nil {
		return template, err
	}

	dataset, err := utils.GetDataset(info.client, info.name, info.namespace)
	if err != nil {
		return template, err
	}

	ownerReference := metav1.OwnerReference{
		APIVersion: dataset.APIVersion,
		Kind:       dataset.Kind,
		Name:       dataset.Name,
		UID:        dataset.UID,
	}

	// 1. set the pvc name
	template = &common.FuseInjectionTemplate{
		PVCName: pvcName,
	}

	// 2. set the fuse container
	if len(ds.Spec.Template.Spec.Containers) != 1 {
		return template, fmt.Errorf("the length of containers of fuse %s in namespace %s is not 1", ds.Name, ds.Namespace)
	}
	template.FuseContainer = ds.Spec.Template.Spec.Containers[0]
	template.FuseContainer.Name = common.FuseContainerName

	// template.VolumesToAdd = ds.Spec.Template.Spec.Volumes
	// 3. inject the post start script for fuse container, if configmap doesn't exist, try to create it.
	mountPath, mountType, err := kubeclient.GetMountInfoFromVolumeClaim(info.client, pvcName, info.namespace)
	if err != nil {
		return
	}
	mountPathInContainer, err := kubeclient.GetFuseMountInContainer(mountType, template.FuseContainer.VolumeMounts)
	if err != nil {
		return
	}

	gen := poststart.NewGenerator(types.NamespacedName{
		Name:      info.name,
		Namespace: info.namespace,
	}, mountPathInContainer.MountPath, mountType)
	cm := gen.BuildConfigmap(ownerReference)
	found, err := kubeclient.IsConfigMapExist(info.client, cm.Name, cm.Namespace)
	if err != nil {
		return template, err
	}

	if !found {
		err = info.client.Create(context.TODO(), cm)
		if err != nil {
			return template, err
		}
	}

	template.FuseContainer.VolumeMounts = append(template.FuseContainer.VolumeMounts, gen.GetVolumeMount())
	if template.FuseContainer.Lifecycle == nil {
		template.FuseContainer.Lifecycle = &corev1.Lifecycle{}
	}
	template.FuseContainer.Lifecycle.PostStart = gen.GetPostStartCommand()

	// 4. create a volume with pvcName with mountpath in pv, and add it to VolumesToUpdate
	template.VolumesToUpdate = []corev1.Volume{
		{
			Name: pvcName,
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: mountPath,
				},
			},
		},
	}
	template.VolumesToAdd = append(ds.Spec.Template.Spec.Volumes, gen.GetVolume())

	return
}

func (info *RuntimeInfo) getFuseDaemonset() (ds *appsv1.DaemonSet, err error) {
	if info.client == nil {
		err = fmt.Errorf("client is not set")
		return
	}

	chartName := ""
	switch info.runtimeType {
	case common.JindoRuntime:
		chartName = common.JindoChartName
	default:
		chartName = info.runtimeType
	}

	fuseName := info.name + "-" + chartName + "-fuse"
	ds, err = kubeclient.GetDaemonset(info.client, fuseName, info.GetNamespace())
	return
}
