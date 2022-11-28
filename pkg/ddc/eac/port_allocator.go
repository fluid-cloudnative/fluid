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

import "github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"

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
	value.Fuse.Port.Monitor = allocatedPorts[index]

	return nil
}

func (e *EACEngine) generateMasterStaticPorts(value *EAC) {
}

func (e *EACEngine) generateWorkerStaticPorts(value *EAC) {
	value.Worker.Port.Rpc = 14555
}

func (e *EACEngine) generateFuseStaticPorts(value *EAC) {
	value.Fuse.Port.Monitor = 15000
}
