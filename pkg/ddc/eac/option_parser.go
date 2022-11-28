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
	"k8s.io/apimachinery/pkg/api/resource"
	"strconv"
	"strings"
)

var (
	miniWorkerQuota = resource.MustParse("1G")
)

func (e *EACEngine) transformMasterOptions(runtime *datav1alpha1.EACRuntime,
	value *EAC) (err error) {

	var (
		masterName   = e.namespace + "-" + e.name + "-master"
		commonOption = "client_owner=" + masterName + ",assign_uuid=" + masterName + ","
		cacheOption  = ""
		portOption   = ""
	)

	commonOption += "g_tier_EnableDadi=true,g_tier_DadiEnablePrefetch=true,"

	value.Master.Options += commonOption
	value.Master.Options += cacheOption
	// TODO: set portOption according to master port
	value.Master.Options += portOption

	for o := range runtime.Spec.Master.Properties {
		option := o + "=" + runtime.Spec.Master.Properties[o] + ","
		value.Master.Options += option
	}

	value.Master.Options = strings.TrimSuffix(value.Master.Options, ",")
	return nil
}

func (e *EACEngine) transformFuseOptions(runtime *datav1alpha1.EACRuntime,
	value *EAC) (err error) {

	var (
		fuseName     = e.namespace + "-" + e.name + "-fuse"
		commonOption = "assign_uuid=" + fuseName + ","
		cacheOption  = ""
		portOption   = ""
	)

	commonOption += "g_tier_EnableDadi=true,g_tier_DadiEnablePrefetch=true,"

	value.Fuse.Options += commonOption
	value.Fuse.Options += cacheOption
	// TODO: set portOption according to fuse port
	value.Fuse.Options += portOption

	for o := range runtime.Spec.Fuse.Properties {
		option := o + "=" + runtime.Spec.Fuse.Properties[o] + ","
		value.Fuse.Options += option
	}

	value.Fuse.Options = strings.TrimSuffix(value.Fuse.Options, ",")
	return nil
}

func (e *EACEngine) transformWorkerOptions(runtime *datav1alpha1.EACRuntime,
	value *EAC) (err error) {

	var (
		commonOption = ""
		cacheOption  = ""
		portOption   = ""
	)

	if value.Worker.TieredStore.Levels[0].Quota != "" {
		quota := resource.MustParse(strings.TrimSuffix(value.Worker.TieredStore.Levels[0].Quota, "B"))
		if miniWorkerQuota.Cmp(quota) > 0 {
			return errors.New(fmt.Sprintf("minimum worker tired store size is %s, current size is %s, please increase size.", miniWorkerQuota.String(), quota.String()))
		}
		quotaValue := quota.Value() / miniWorkerQuota.Value()
		cacheOption += "cache_capacity_gb=" + strconv.Itoa(int(quotaValue)) + ","
	}

	if value.Worker.TieredStore.Levels[0].MediumType == string(common.Memory) {
		cacheOption += "tmpfs=true,"
	}

	cacheOption += "cache_media=" + value.getTiredStoreLevel0Path() + ","
	portOption += "server_port=" + strconv.Itoa(value.Worker.Port.Rpc) + ","

	value.Worker.Options += commonOption
	value.Worker.Options += cacheOption
	value.Worker.Options += portOption

	for o := range runtime.Spec.Worker.Properties {
		option := o + "=" + runtime.Spec.Worker.Properties[o] + ","
		value.Worker.Options += option
	}

	value.Worker.Options = strings.TrimSuffix(value.Worker.Options, ",")
	return nil
}
