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
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
)

func (e *EACEngine) transformPortForMaster(runtime *datav1alpha1.EACRuntime, value *EAC) (err error) {
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

func (e *EACEngine) transformPortForWorker(runtime *datav1alpha1.EACRuntime, value *EAC) (err error) {
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

func (e *EACEngine) transformPortForFuse(runtime *datav1alpha1.EACRuntime, value *EAC) (err error) {
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

func (e *EACEngine) allocateMasterPorts(value *EAC) error {
	return nil
}

func (e *EACEngine) allocateWorkerPorts(value *EAC) error {
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

func (e *EACEngine) allocateFusePorts(value *EAC) error {
	return nil
}

func (e *EACEngine) generateMasterStaticPorts(value *EAC) {
}

func (e *EACEngine) generateWorkerStaticPorts(value *EAC) {
	value.Worker.Port.Rpc = 14555
}

func (e *EACEngine) generateFuseStaticPorts(value *EAC) {
}
