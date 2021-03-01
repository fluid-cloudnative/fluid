/*

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

package base

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	v1 "k8s.io/api/batch/v1"
)

// Load the data and return DataLoad job status
func (t *TemplateEngine) LoadData(ctx cruntime.ReconcileRequestContext, targetDataload datav1alpha1.DataLoad) (status v1.JobConditionType, err error) {
	if err = t.Implement.CreateDataLoadJob(ctx, targetDataload); err != nil {
		t.Log.Error(err, "Failed to create the DataLoad job.")
		return "", err
	}
	status, err = t.Implement.GetDataLoadJobStatus(ctx, targetDataload)
	if err != nil {
		t.Log.Error(err, "Failed to check whether the DataLoad job is finished or not")
		return "", err
	}
	return status, nil
}

// Check if the runtime is ready
func (t *TemplateEngine) Ready() (ready bool, err error) {
	masterReady, err := t.Implement.CheckMasterReady()
	if err != nil {
		t.Log.Error(err, "Failed to check if the master is ready.")
		return ready, err
	}
	workersReady, err := t.Implement.CheckWorkersReady()
	if err != nil {
		t.Log.Error(err, "Failed to check if the workers are ready.")
		return ready, err
	}
	ready = masterReady && workersReady
	return ready, nil
}
