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

package thin

import datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"

func (t *ThinEngine) transformConfig(runtime *datav1alpha1.ThinRuntime,
	profile *datav1alpha1.ThinRuntimeProfile,
	dataset *datav1alpha1.Dataset) (config Config, err error) {
	mounts := []datav1alpha1.Mount{}
	for _, m := range dataset.Spec.Mounts {
		m.Options, err = t.genUFSMountOptions(m)
		if err != nil {
			return
		}
		m.EncryptOptions = []datav1alpha1.EncryptOption{}
		mounts = append(mounts, m)
	}

	config.Mounts = mounts
	config.RuntimeOptions = runtime.Spec.Fuse.Options
	return
}

func (t *ThinEngine) initRuntimeStatus(runtime *datav1alpha1.ThinRuntime,
	profile *datav1alpha1.ThinRuntimeProfile,
	dataset *datav1alpha1.Dataset) (status RuntimeStatus) {

	status.Workers = []string{}

	return RuntimeStatus{
		Workers: []string{},
		Fuses:   []string{},
	}
}
