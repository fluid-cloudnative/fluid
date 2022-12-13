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
	"github.com/pkg/errors"
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
func GetMountInfoFromVolumeClaim(client client.Client, name, namespace string) (path string, mountType string, subpath string, err error) {
	pvc, err := GetPersistentVolumeClaim(client, name, namespace)
	if err != nil {
		err = errors.Wrapf(err, "failed to get persistent volume claim")
		return
	}

	pv, err := GetPersistentVolume(client, pvc.Spec.VolumeName)
	if err != nil {
		err = errors.Wrapf(err, "cannot find pvc \"%s/%s\"'s bounded PV", pvc.Namespace, pvc.Name)
		return
	}

	if pv.Spec.CSI != nil && len(pv.Spec.CSI.VolumeAttributes) > 0 {
		path = pv.Spec.CSI.VolumeAttributes[common.VolumeAttrFluidPath]
		mountType = pv.Spec.CSI.VolumeAttributes[common.VolumeAttrMountType]
		subpath = pv.Spec.CSI.VolumeAttributes[common.VolumeAttrFluidSubPath]
	} else {
		err = fmt.Errorf("the pvc %s in %s is not created by fluid",
			name,
			namespace)
	}

	return
}
