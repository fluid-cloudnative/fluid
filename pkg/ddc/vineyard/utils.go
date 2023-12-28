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

package vineyard

import (
	"context"
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

// getRuntime gets the vineyard runtime
func (e *VineyardEngine) getRuntime() (*datav1alpha1.VineyardRuntime, error) {
	key := types.NamespacedName{
		Name:      e.name,
		Namespace: e.namespace,
	}

	var runtime datav1alpha1.VineyardRuntime
	if err := e.Get(context.TODO(), key, &runtime); err != nil {
		return nil, err
	}
	return &runtime, nil
}

func (e *VineyardEngine) getMasterPod(name string, namespace string) (pod *v1.Pod, err error) {
	pod = &v1.Pod{}
	err = e.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, pod)

	return pod, err
}

func (e *VineyardEngine) getMasterStatefulset(name string, namespace string) (master *appsv1.StatefulSet, err error) {
	master = &appsv1.StatefulSet{}
	err = e.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, master)

	return master, err
}

func (e *VineyardEngine) getDaemonset(name string, namespace string) (daemonset *appsv1.DaemonSet, err error) {
	daemonset = &appsv1.DaemonSet{}
	err = e.Client.Get(context.TODO(), types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, daemonset)

	return daemonset, err
}

func (e *VineyardEngine) getWorkerPodExporterPort() (port int32) {
	ports := e.runtime.Spec.Worker.Ports
	if len(ports) == 0 {
		port = WorkerExporterPort
	} else {
		for n, p := range ports {
			if n == WorkerExporterName {
				port = int32(p)
				break
			}
		}
	}

	return
}

func (e *VineyardEngine) getWorkerReplicas() (replicas int32) {
	replicas = e.runtime.Spec.Worker.Replicas
	if replicas == 0 {
		replicas = 1
	}
	return
}

func (e *VineyardEngine) getMasterName() (dsName string) {
	return e.name + "-master"
}

func (e *VineyardEngine) getWorkerName() (dsName string) {
	return e.name + "-worker"
}

func (e *VineyardEngine) getFuseDaemonsetName() (dsName string) {
	return e.name + "-fuse"
}

func (e *VineyardEngine) getMountPoint() (mountPath string) {
	mountRoot := getMountRoot()
	e.Log.Info("mountRoot", "path", mountRoot)
	return fmt.Sprintf("%s/%s/%s/vineyard-fuse", mountRoot, e.namespace, e.name)
}

func (e *VineyardEngine) getTargetPath() (targetPath string) {
	mountRoot := getMountRoot()
	e.Log.Info("mountRoot", "path", mountRoot)
	return fmt.Sprintf("%s/%s/%s/vineyard-fuse", mountRoot, e.namespace, e.name)
}

// getMountRoot returns the default path, if it's not set
func getMountRoot() (path string) {
	path, err := utils.GetMountRoot()
	if err != nil {
		path = "/" + common.VineyardRuntime
	} else {
		path = path + "/" + common.VineyardRuntime
	}
	return
}

func (e *VineyardEngine) parseMasterImage(image string, tag string, imagePullPolicy string) (string, string, string) {
	if len(imagePullPolicy) == 0 {
		imagePullPolicy = common.DefaultImagePullPolicy
	}

	if len(image) == 0 {
		masterImageInfo := strings.Split(common.DefaultVineyardMasterImage, ":")
		if len(masterImageInfo) < 1 {
			panic("invalid default vineyard master image!")
		} else {
			image = masterImageInfo[0]
		}
	}

	if len(tag) == 0 {
		masterImageInfo := strings.Split(common.DefaultVineyardMasterImage, ":")
		if len(masterImageInfo) < 2 {
			panic("invalid default vineyard master image!")
		} else {
			tag = masterImageInfo[1]
		}
	}

	return image, tag, imagePullPolicy
}

func (e *VineyardEngine) parseWorkerImage(image string, tag string, imagePullPolicy string) (string, string, string) {
	if len(imagePullPolicy) == 0 {
		imagePullPolicy = common.DefaultImagePullPolicy
	}

	if len(image) == 0 {
		workerImageInfo := strings.Split(common.DefaultVineyardWorkerImage, ":")
		if len(workerImageInfo) < 1 {
			panic("invalid default vineyard worker image!")
		} else {
			image = workerImageInfo[0]
		}
	}

	if len(tag) == 0 {
		workerImageInfo := strings.Split(common.DefaultVineyardWorkerImage, ":")
		if len(workerImageInfo) < 2 {
			panic("invalid default vineyard worker image!")
		} else {
			tag = workerImageInfo[1]
		}
	}

	return image, tag, imagePullPolicy
}

func (e *VineyardEngine) parseFuseImage(image string, tag string, imagePullPolicy string) (string, string, string) {
	if len(imagePullPolicy) == 0 {
		imagePullPolicy = common.DefaultImagePullPolicy
	}

	if len(image) == 0 {
		fuseImageInfo := strings.Split(common.DefultVineyardFuseImage, ":")
		if len(fuseImageInfo) < 1 {
			panic("invalid default vineyard fuse image!")
		} else {
			image = fuseImageInfo[0]
		}
	}

	if len(tag) == 0 {
		fuseImageInfo := strings.Split(common.DefultVineyardFuseImage, ":")
		if len(fuseImageInfo) < 2 {
			panic("invalid default vineyard fuse image!")
		} else {
			tag = fuseImageInfo[1]
		}
	}

	return image, tag, imagePullPolicy
}
