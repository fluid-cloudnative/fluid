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

package kubeclient

import (
	"context"
	"fmt"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetPersistentVolumeClaim(client client.Client, name, namespace string) (pvc *v1.PersistentVolumeClaim, err error) {
	pvc = &v1.PersistentVolumeClaim{}
	err = client.Get(context.TODO(),
		types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
		pvc)
	return
}

// GetMountInfoFromVolumeClaim gets the mountPath and type for CSI plugin
func GetMountInfoFromVolumeClaim(client client.Client, name, namespace string) (path string, mountType string, err error) {
	pvc, err := GetPersistentVolumeClaim(client, name, namespace)
	if err != nil {
		return
	}

	pv, err := GetPersistentVolume(client, pvc.Spec.VolumeName)
	if err != nil {
		return
	}

	if pv.Spec.CSI != nil && len(pv.Spec.CSI.VolumeAttributes) > 0 {
		path = pv.Spec.CSI.VolumeAttributes[common.FluidPath]
		mountType = pv.Spec.CSI.VolumeAttributes[common.MountType]
	} else {
		err = fmt.Errorf("the pvc %s in %s is not created by fluid",
			name,
			namespace)
	}

	return
}
