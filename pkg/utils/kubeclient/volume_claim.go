/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
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
