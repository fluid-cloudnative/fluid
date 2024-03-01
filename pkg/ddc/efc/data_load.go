/*
  Copyright 2023 The Fluid Authors.

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
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cdataload "github.com/fluid-cloudnative/fluid/pkg/dataload"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/efc/operations"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	v1 "k8s.io/api/core/v1"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
)

// CreateDataLoadJob creates the job to load data
func (e *EFCEngine) CreateDataLoadJob(ctx cruntime.ReconcileRequestContext, targetDataload datav1alpha1.DataLoad) (err error) {
	log := ctx.Log.WithName("createDataLoadJob")

	// 1. Check if the helm release already exists
	releaseName := utils.GetDataLoadReleaseName(targetDataload.Name)
	jobName := utils.GetDataLoadJobName(releaseName)
	var existed bool
	existed, err = helm.CheckRelease(releaseName, targetDataload.Namespace)
	if err != nil {
		log.Error(err, "failed to check if release exists", "releaseName", releaseName, "namespace", targetDataload.Namespace)
		return err
	}

	// 2. install the helm chart if not exists
	if !existed {
		log.Info("DataLoad job helm chart not installed yet, will install")
		valueFileName, err := e.generateDataLoadValueFile(ctx, targetDataload)
		if err != nil {
			log.Error(err, "failed to generate dataload chart's value file")
			return err
		}
		chartName := utils.GetChartsDirectory() + "/" + cdataload.DataloadChart + "/" + common.EFCRuntime
		err = helm.InstallRelease(releaseName, targetDataload.Namespace, valueFileName, chartName)
		if err != nil {
			log.Error(err, "failed to install dataload chart")
			return err
		}
		log.Info("DataLoad job helm chart successfully installed", "namespace", targetDataload.Namespace, "releaseName", releaseName)
		ctx.Recorder.Eventf(&targetDataload, v1.EventTypeNormal, common.DataLoadJobStarted, "The DataLoad job %s started", jobName)
	}
	return err
}

// CheckRuntimeReady checks if runtime is ready or not
func (e *EFCEngine) CheckRuntimeReady() (ready bool) {
	// 1. check master ready
	podName, containerName := e.getMasterPodInfo()
	fileUtils := operations.NewEFCFileUtils(podName, containerName, e.namespace, e.Log)
	ready = fileUtils.Ready()
	if !ready {
		e.Log.Info("runtime not ready", "runtime", ready)
		return false
	}

	// 2. check worker ready
	workerPods, err := e.getWorkerRunningPods()
	if err != nil {
		e.Log.Error(err, "Fail to get worker pods")
		return false
	}

	readyCount := 0
	for _, pod := range workerPods {
		if podutil.IsPodReady(&pod) {
			readyCount++
		}
	}
	return readyCount > 0
}

// generateDataLoadValueFile builds a DataLoadValue by extracted specifications from the given DataLoad, and
// marshals the DataLoadValue to a temporary yaml file where stores values that'll be used by fluid dataloader helm chart
func (e *EFCEngine) generateDataLoadValueFile(r cruntime.ReconcileRequestContext, dataload datav1alpha1.DataLoad) (valueFileName string, err error) {
	// TODO: generate dataload value file
	return "", nil
}
