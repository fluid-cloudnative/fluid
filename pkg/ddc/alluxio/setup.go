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

package alluxio

import (
	datav1alpha1 "github.com/cloudnativefluid/fluid/api/v1alpha1"
	"github.com/cloudnativefluid/fluid/pkg/utils"
)

// IsSetupDone checks the setup is done
func (e *AlluxioEngine) IsSetupDone() (done bool, err error) {

	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		return done, err
	}

	// If the dataset condition is created, it means the dataset is already setup
	index, _ := utils.GetDatasetCondition(dataset.Status.Conditions, datav1alpha1.DatasetReady)
	if index != -1 {
		e.Log.V(1).Info("The runtime is already setup.")
		done = true
		return done, nil
	}

	return
}
