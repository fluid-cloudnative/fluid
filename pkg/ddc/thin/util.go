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

import (
	"context"
	"fmt"
	"strconv"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
	options "sigs.k8s.io/controller-runtime/pkg/client"
)

// getRuntime gets thin runtime
func (t *ThinEngine) getRuntime() (*datav1alpha1.ThinRuntime, error) {

	key := types.NamespacedName{
		Name:      t.name,
		Namespace: t.namespace,
	}

	var runtime datav1alpha1.ThinRuntime
	if err := t.Get(context.TODO(), key, &runtime); err != nil {
		return nil, err
	}
	return &runtime, nil
}

func (t *ThinEngine) getThinRuntimeProfile() (*datav1alpha1.ThinRuntimeProfile, error) {
	if t.runtime == nil {
		return nil, nil
	}
	key := types.NamespacedName{
		Name: t.runtime.Spec.ThinRuntimeProfileName,
	}

	var profile datav1alpha1.ThinRuntimeProfile
	if err := t.Get(context.TODO(), key, &profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

func (t *ThinEngine) getFuseName() (dsName string) {
	return t.name + "-fuse"
}

func (t *ThinEngine) getWorkerName() (dsName string) {
	return t.name + "-worker"
}

func (t *ThinEngine) getTargetPath() (targetPath string) {
	mountRoot := getMountRoot()
	t.Log.Info("mountRoot", "path", mountRoot)
	return fmt.Sprintf("%s/%s/%s/thin-fuse", mountRoot, t.namespace, t.name)
}

// getMountRoot returns the default path, if it's not set
func getMountRoot() (path string) {
	path, err := utils.GetMountRoot()
	if err != nil {
		path = "/" + common.ThinRuntime
	} else {
		path = path + "/" + common.ThinRuntime
	}
	return
}

func (t *ThinEngine) getDaemonset(name string, namespace string) (fuse *appsv1.DaemonSet, err error) {
	fuse = &appsv1.DaemonSet{}
	err = t.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, fuse)

	return fuse, err
}

func (t *ThinEngine) GetRunningPodsOfDaemonset(dsName string, namespace string) (pods []corev1.Pod, err error) {
	ds, err := t.getDaemonset(dsName, namespace)
	if err != nil {
		return pods, err
	}

	selector := ds.Spec.Selector.MatchLabels

	pods = []corev1.Pod{}
	podList := &corev1.PodList{}
	err = t.Client.List(context.TODO(), podList, options.InNamespace(namespace), options.MatchingLabels(selector))
	if err != nil {
		return pods, err
	}

	for _, pod := range podList.Items {
		if !podutil.IsPodReady(&pod) {
			t.Log.Info("Skip the pod because it's not ready", "pod", pod.Name, "namespace", pod.Namespace)
			continue
		}
		pods = append(pods, pod)
	}

	return pods, nil
}

func (t *ThinEngine) getDataSetFileNum() (string, error) {
	fileCount, err := t.TotalFileNums()
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(fileCount, 10), err
}

func (t ThinEngine) getFuseConfigMapName() string {
	return t.name + "-fuse-conf"
}
