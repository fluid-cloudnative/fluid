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

// SetupWorkers checks the desired and current replicas of workers and makes an update
// over the status by setting phases and conditions. The function
// calls for a status update and finally returns error if anything unexpected happens.
func (e *JindoEngine) SetupWorkers() (err error) {
	return
}

// ShouldSetupWorkers checks if we need setup the workers
func (e *JindoEngine) ShouldSetupWorkers() (should bool, err error) {
	return true, nil
}

// are the workers ready
func (e *JindoEngine) CheckWorkersReady() (ready bool, err error) {
	return true, nil
}
