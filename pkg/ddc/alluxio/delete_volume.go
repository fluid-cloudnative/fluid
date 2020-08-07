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
	"fmt"
	"time"

	"github.com/cloudnativefluid/fluid/pkg/utils/kubeclient"
)

// DeleteVolume creates volume
func (e *AlluxioEngine) DeleteVolume() (err error) {

	if e.runtime == nil {
		e.runtime, err = e.getRuntime()
		if err != nil {
			return
		}
	}

	err = e.deleteFusePersistentVolumeClaim()
	if err != nil {
		return
	}

	err = e.deleteFusePersistentVolume()
	if err != nil {
		return
	}

	return

}

// deleteFusePersistentVolume
func (e *AlluxioEngine) deleteFusePersistentVolume() (err error) {

	found, err := kubeclient.IsPersistentVolumeExist(e.Client, e.runtime.Name, expectedAnnotations)
	if err != nil {
		return err
	}

	if found {
		err = kubeclient.DeletePersistentVolume(e.Client, e.runtime.Name)
		if err != nil {
			return err
		}
		retries := 500
		for i := 0; i < retries; i++ {
			found, err = kubeclient.IsPersistentVolumeExist(e.Client, e.runtime.Name, expectedAnnotations)
			if err != nil {
				return err
			}

			if found {
				time.Sleep(time.Duration(2 * time.Second))
			} else {
				break
			}
		}

		if found {
			return fmt.Errorf("The PV %s is not cleaned up",
				e.runtime.Name)
		} else {
			e.Log.Info("The PV is deleted successfully", "name", e.runtime.Name)
		}
	}

	return err
}

// deleteFusePersistentVolume
func (e *AlluxioEngine) deleteFusePersistentVolumeClaim() (err error) {

	found, err := kubeclient.IsPersistentVolumeClaimExist(e.Client, e.runtime.Name, e.runtime.Namespace, expectedAnnotations)
	if err != nil {
		return err
	}

	if found {
		err = kubeclient.DeletePersistentVolumeClaim(e.Client, e.runtime.Name, e.runtime.Namespace)
		if err != nil {
			return err
		}

		retries := 500
		for i := 0; i < retries; i++ {
			found, err = kubeclient.IsPersistentVolumeClaimExist(e.Client, e.runtime.Name, e.runtime.Namespace, expectedAnnotations)
			if err != nil {
				return err
			}

			if found {
				time.Sleep(time.Duration(2 * time.Second))
			} else {
				break
			}
		}

		if found {
			return fmt.Errorf("The PV %s in ns %s is not cleaned up",
				e.runtime.Name,
				e.runtime.Namespace)
		} else {
			e.Log.Info("The PVC is deleted successfully",
				"name", e.runtime.Name,
				"namespace", e.runtime.Namespace)
		}
	}

	return err

}
