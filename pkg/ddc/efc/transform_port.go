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
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
)

func (e *EFCEngine) transformPortForMaster(runtime *datav1alpha1.EFCRuntime, value *EFC) (err error) {
	if datav1alpha1.IsHostNetwork(runtime.Spec.Master.NetworkMode) {
		e.Log.Info("allocateMasterPorts for hostnetwork mode")
		err = e.allocateMasterPorts(value)
		if err != nil {
			return
		}
	} else {
		e.Log.Info("skip allocateMasterPorts for container network mode")
		e.generateMasterStaticPorts(value)
	}
	return
}

func (e *EFCEngine) transformPortForWorker(runtime *datav1alpha1.EFCRuntime, value *EFC) (err error) {
	if datav1alpha1.IsHostNetwork(runtime.Spec.Worker.NetworkMode) {
		e.Log.Info("allocateWorkerPorts for hostnetwork mode")
		err = e.allocateWorkerPorts(value)
		if err != nil {
			return
		}
	} else {
		e.Log.Info("skip allocateWorkerPorts for container network mode")
		e.generateWorkerStaticPorts(value)
	}
	return
}

func (e *EFCEngine) transformPortForFuse(runtime *datav1alpha1.EFCRuntime, value *EFC) (err error) {
	if datav1alpha1.IsHostNetwork(runtime.Spec.Fuse.NetworkMode) {
		e.Log.Info("allocateFusePorts for hostnetwork mode")
		err = e.allocateFusePorts(value)
		if err != nil {
			return
		}
	} else {
		e.Log.Info("skip allocateFusePorts for container network mode")
		e.generateFuseStaticPorts(value)
	}
	return
}

func (e *EFCEngine) allocateMasterPorts(value *EFC) error {
	return nil
}

func (e *EFCEngine) allocateWorkerPorts(value *EFC) error {
	expectedPortNum := 1

	allocator, err := portallocator.GetRuntimePortAllocator()
	if err != nil {
		e.Log.Error(err, "can't get runtime port allocator")
		return err
	}

	allocatedPorts, err := allocator.GetAvailablePorts(expectedPortNum)
	if err != nil {
		e.Log.Error(err, "can't get available ports", "expected port num", expectedPortNum)
		return err
	}

	index := 0
	value.Worker.Port.Rpc = allocatedPorts[index]

	return nil
}

func (e *EFCEngine) allocateFusePorts(value *EFC) error {
	return nil
}

func (e *EFCEngine) generateMasterStaticPorts(value *EFC) {
}

func (e *EFCEngine) generateWorkerStaticPorts(value *EFC) {
	value.Worker.Port.Rpc = 14555
}

func (e *EFCEngine) generateFuseStaticPorts(value *EFC) {
}
