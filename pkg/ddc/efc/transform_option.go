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

package efc

import (
	"fmt"
	"strings"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"k8s.io/apimachinery/pkg/api/resource"
)

var (
	miniWorkerQuota = resource.MustParse("1Gi")
)

func (e *EFCEngine) transformMasterOptions(runtime *datav1alpha1.EFCRuntime,
	value *EFC, info *MountInfo) (err error) {
	var options []string

	if info.MountPointPrefix == CpfsMountPointPrefix {
		options = append(options, "protocol=nfs3")
	}

	options = append(options, fmt.Sprintf("client_owner=%s-%s-master", e.namespace, e.name))
	options = append(options, fmt.Sprintf("assign_uuid=%s-%s-master", e.namespace, e.name))

	for o := range runtime.Spec.Master.Properties {
		options = append(options, fmt.Sprintf("%s=%s", o, runtime.Spec.Master.Properties[o]))
	}

	value.Master.Options = strings.Join(options, ",")
	return nil
}

func (e *EFCEngine) transformFuseOptions(runtime *datav1alpha1.EFCRuntime,
	value *EFC, info *MountInfo) (err error) {
	var options []string

	if info.MountPointPrefix == CpfsMountPointPrefix {
		options = append(options, "protocol=nfs3")
	}

	if !runtime.Spec.Worker.Disabled {
		options = append(options, "g_tier_EnableClusterCache=true")
		options = append(options, "g_tier_EnableClusterCachePrefetch=true")
	}

	options = append(options, fmt.Sprintf("assign_uuid=%s-%s-fuse", e.namespace, e.name))

	for o := range runtime.Spec.Fuse.Properties {
		options = append(options, fmt.Sprintf("%s=%s", o, runtime.Spec.Fuse.Properties[o]))
	}

	value.Fuse.Options = strings.Join(options, ",")
	return nil
}

func (e *EFCEngine) transformWorkerOptions(runtime *datav1alpha1.EFCRuntime,
	value *EFC) (err error) {
	if len(value.Worker.TieredStore.Levels) == 0 {
		return fmt.Errorf("worker tiered store are not specified")
	}

	var options []string
	options = append(options, fmt.Sprintf("cache_media=%s", value.getTiredStoreLevel0Path()))
	options = append(options, fmt.Sprintf("server_port=%v", value.Worker.Port.Rpc))

	if len(value.getTiredStoreLevel0Quota()) > 0 {
		quota := *utils.TransformEFCUnitToQuantity(value.getTiredStoreLevel0Quota())
		if miniWorkerQuota.Cmp(quota) > 0 {
			return fmt.Errorf("minimum worker tired store size is %s, current size is %s, please increase size", miniWorkerQuota.String(), quota.String())
		}
		quotaValue := quota.Value() / miniWorkerQuota.Value()
		options = append(options, fmt.Sprintf("cache_capacity_gb=%v", int(quotaValue)))
	}
	if value.getTiredStoreLevel0MediumType() == string(common.Memory) {
		options = append(options, "tmpfs=true")
	}

	for o := range runtime.Spec.Worker.Properties {
		options = append(options, fmt.Sprintf("%s=%s", o, runtime.Spec.Worker.Properties[o]))
	}

	value.Worker.Options = strings.Join(options, ",")
	return nil
}
