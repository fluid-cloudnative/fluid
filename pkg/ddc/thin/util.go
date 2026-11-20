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
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
	options "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
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

// GetValuesConfigMap returns the ConfigMap holding the Helm values that the runtime was last
// rendered with. It returns a nil ConfigMap without an error when the ConfigMap does not exist,
// which happens when the user opted out via common.AnnotationDisableRuntimeHelmValueConfig.
func (t *ThinEngine) GetValuesConfigMap() (cm *corev1.ConfigMap, err error) {
	cm = &corev1.ConfigMap{}
	err = t.Client.Get(context.TODO(), types.NamespacedName{
		Name:      t.getHelmValuesConfigMapName(),
		Namespace: t.namespace,
	}, cm)
	if apierrs.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return cm, nil
}

// GetValueFromConfigmap returns the last synced ThinValue. A nil value without an error means the
// values ConfigMap is not available, so there is no state to sync against.
func (t *ThinEngine) GetValueFromConfigmap() (*ThinValue, error) {
	cm, err := t.GetValuesConfigMap()
	if err != nil || cm == nil {
		return nil, err
	}

	data, exist := cm.Data["data"]
	if !exist {
		return nil, fmt.Errorf("no data key found in the helm value configmap %s/%s", t.namespace, cm.Name)
	}

	var value ThinValue
	if err := yaml.Unmarshal([]byte(data), &value); err != nil {
		return nil, err
	}

	return &value, nil
}

// SaveValueToConfigmap persists the value as the new last synced state.
func (t *ThinEngine) SaveValueToConfigmap(value *ThinValue) error {
	data, err := yaml.Marshal(value)
	if err != nil {
		return err
	}

	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		cm, err := t.GetValuesConfigMap()
		if err != nil {
			return err
		}
		if cm == nil {
			return fmt.Errorf("helm value configmap %s/%s not found", t.namespace, t.getHelmValuesConfigMapName())
		}

		if cm.Data == nil {
			cm.Data = map[string]string{}
		}
		cm.Data["data"] = string(data)
		return kubeclient.UpdateConfigMap(t.Client, cm)
	})
}

func (t ThinEngine) isWorkerEnable() bool {
	if t.runtime == nil {
		return false
	}
	return t.runtime.Spec.Worker.Enabled
}
