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
	"context"
	"fmt"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

type VineyardMutator struct {
	// VineyardMutator inherits from DefaultMutator
	DefaultMutator
}

var _ Mutator = &VineyardMutator{}

func NewVineyardMutator(opts MutatorBuildOpts) Mutator {
	return &VineyardMutator{
		DefaultMutator: DefaultMutator{
			options: opts.Options,
			client:  opts.Client,
			log:     opts.Log,
			Specs:   opts.Specs,
		},
	}
}

func (mutator *VineyardMutator) MutateWithRuntimeInfo(pvcName string, runtimeInfo base.RuntimeInfoInterface, nameSuffix string) error {
	// check if the runtime is Vineyard
	if runtimeInfo.GetRuntimeType() != common.VineyardRuntime {
		return fmt.Errorf("runtime %s is not %s", runtimeInfo.GetName(), common.VineyardRuntime)
	}

	// check whether the vineyard rpc configmap is existed
	cm := &corev1.ConfigMap{}
	cmName := pvcName + common.VineyardConfigmapSuffix
	cmNamespace := runtimeInfo.GetNamespace()
	if err := mutator.client.Get(context.TODO(), types.NamespacedName{Name: cmName, Namespace: cmNamespace}, cm); err != nil {
		return err
	}

	// add vineyard configmap volume
	volumes := mutator.Specs.Volumes
	overriddenVolumes := make([]corev1.Volume, 0)
	pvcVolumeName := ""
	for _, volume := range volumes {
		switch {
		case volume.Name == common.VineyardConfigmapVolumeName:
			continue
		case volume.PersistentVolumeClaim != nil && volume.PersistentVolumeClaim.ClaimName == pvcName:
			pvcVolumeName = volume.Name
			continue
		}
		overriddenVolumes = append(overriddenVolumes, volume)
	}

	overriddenVolumes = append(overriddenVolumes, corev1.Volume{
		Name: common.VineyardConfigmapVolumeName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: cmName,
				},
			},
		},
	})

	mutator.Specs.Volumes = overriddenVolumes

	// inject vineyard volume as environment variable
	containers := mutator.Specs.Containers
	for i := range containers {
		overriddenEnvs := make([]corev1.EnvVar, 0)
		for j := range containers[i].Env {
			if containers[i].Env[j].Name != common.VineyardConfigmapVolumeName {
				overriddenEnvs = append(overriddenEnvs, containers[i].Env[j])
			}
		}
		overriddenEnvs = append(overriddenEnvs, corev1.EnvVar{
			Name: common.VineyardRPCEndpoint,
			ValueFrom: &corev1.EnvVarSource{
				ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: cmName,
					},
					Key: common.VineyardRPCEndpoint,
				},
			},
		})
		overridenVolumeMounts := make([]corev1.VolumeMount, 0)
		for _, vm := range containers[i].VolumeMounts {
			if pvcVolumeName != "" && vm.Name != pvcVolumeName {
				overridenVolumeMounts = append(overridenVolumeMounts, vm)
			}
		}

		containers[i].Env = overriddenEnvs
		containers[i].VolumeMounts = overridenVolumeMounts
	}
	mutator.Specs.Containers = containers

	volumesMounts := mutator.Specs.VolumeMounts
	overridenVolumeMounts := make([]corev1.VolumeMount, 0)
	for _, volumeMount := range volumesMounts {
		if pvcVolumeName != "" && volumeMount.Name != pvcVolumeName {
			overridenVolumeMounts = append(overridenVolumeMounts, volumeMount)
		}
	}
	mutator.Specs.VolumeMounts = overridenVolumeMounts
	return nil
}

func (mutator *VineyardMutator) PostMutate() error {
	return mutator.DefaultMutator.PostMutate()
}

func (mutator *VineyardMutator) GetMutatedPodSpecs() *MutatingPodSpecs {
	return mutator.DefaultMutator.GetMutatedPodSpecs()
}
