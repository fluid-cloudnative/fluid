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
