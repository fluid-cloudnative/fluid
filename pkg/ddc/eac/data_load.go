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
	"github.com/fluid-cloudnative/fluid/pkg/ddc/eac/operations"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/pkg/errors"
	"path/filepath"
)

// CreateDataLoadJob creates the job to load data
func (e *EACEngine) CreateDataLoadJob(ctx cruntime.ReconcileRequestContext, targetDataload datav1alpha1.DataLoad) (err error) {
	return errors.New("EAC currently not support load data")
}

func (e *EACEngine) CheckRuntimeReady() (ready bool) {
	podName, containerName := e.getMasterPodInfo()
	fileUtils := operations.NewEACFileUtils(podName, containerName, e.namespace, e.Log)
	ready = fileUtils.Ready()
	if !ready {
		e.Log.Info("runtime not ready", "runtime", ready)
		return false
	}
	return true
}

func (e *EACEngine) CheckExistenceOfPath(targetDataload datav1alpha1.DataLoad) (notExist bool, err error) {
	podName, containerName := e.getMasterPodInfo()
	fileUtils := operations.NewEACFileUtils(podName, containerName, e.namespace, e.Log)

	for _, target := range targetDataload.Spec.Target {
		targetPath := filepath.Join(MasterMountPath, target.Path)
		isExist, err := fileUtils.IsExist(targetPath)
		if err != nil {
			return true, err
		}
		if !isExist {
			return true, nil
		}
	}
	return false, nil
}
