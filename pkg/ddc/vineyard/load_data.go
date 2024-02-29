/*
Copyright 2024 The Fluid Authors.
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

package vineyard

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
)

// CreateDataLoadJob creates the job to load data
func (e *VineyardEngine) CreateDataLoadJob(ctx cruntime.ReconcileRequestContext, targetDataload datav1alpha1.DataLoad) (err error) {
	// TODO:(caoye) implement me
	return err
}

func (e *VineyardEngine) CheckRuntimeReady() (ready bool) {
	// TODO:(caoye) implement me
	return true
}
