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
	"github.com/cloudnativefluid/fluid/pkg/ddc/alluxio/operations"
)

func (e *AlluxioEngine) usedStorageBytesInternal() (value int64, err error) {
	return
}

func (e *AlluxioEngine) freeStorageBytesInternal() (value int64, err error) {
	return
}

func (e *AlluxioEngine) totalStorageBytesInternal() (total int64, err error) {
	podName, containerName := e.getMasterPodInfo()

	fileUitls := operations.NewAlluxioFileUtils(podName, containerName, e.namespace, e.Log)
	_, _, total, err = fileUitls.Count("/")
	if err != nil {
		return
	}

	return
}

func (e *AlluxioEngine) totalFileNumsInternal() (fileCount int64, err error) {
	podName, containerName := e.getMasterPodInfo()

	fileUitls := operations.NewAlluxioFileUtils(podName, containerName, e.namespace, e.Log)
	fileCount, _, _, err = fileUitls.Count("/")
	if err != nil {
		return
	}

	return
}
