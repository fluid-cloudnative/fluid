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

package jindo

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	batchv1 "k8s.io/api/batch/v1"
)

// CreateDataLoadJob creates the job to load data
func (e *JindoEngine) CreateDataLoadJob(ctx cruntime.ReconcileRequestContext, targetDataload datav1alpha1.DataLoad) (err error) {
	// todo
	return err
}

// GetDataLoadJobStatus checks whether the DataLoad job is finished or not
func (e *JindoEngine) GetDataLoadJobStatus(ctx cruntime.ReconcileRequestContext, targetDataload datav1alpha1.DataLoad) (status batchv1.JobConditionType, err error) {
	// todo
	return status, err
}
