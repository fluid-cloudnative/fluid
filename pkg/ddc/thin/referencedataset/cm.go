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

package referencedataset

import (
	"context"
	"fmt"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func createConfigMapForRefDataset(client client.Client, refDataset *datav1alpha1.Dataset, mountedRuntime base.RuntimeInfoInterface) error {
	mountedRuntimeType := mountedRuntime.GetRuntimeType()
	mountedRuntimeName := mountedRuntime.GetName()
	mountedRuntimeNamespace := mountedRuntime.GetNamespace()

	refNameSpace := refDataset.GetNamespace()

	// add owner reference to ensure config map deleted when delete the dataset
	ownerReference := metav1.OwnerReference{
		APIVersion: refDataset.APIVersion,
		Kind:       refDataset.Kind,
		Name:       refDataset.Name,
		UID:        refDataset.UID,
	}

	// copy the configmap to ref namespace.
	// TODO: any other config resource like secret need copied ?

	// Note: values configmap is not needed for fuse sidecar container.

	// TODO: decoupling the switch-case, too fragile
	switch mountedRuntimeType {
	// TODO:  currently the dst configmap name is the same as src configmap name to avoid modify the fuse init container filed,
	//       but duplicated name error can occurs if the dst namespace has same named runtime.
	case common.AlluxioRuntime:
		configMapName := mountedRuntimeName + "-config"
		err := copyConfigMap(client, configMapName, mountedRuntimeNamespace, configMapName, refNameSpace, ownerReference)
		if err != nil {
			return err
		}
	case common.JuiceFSRuntime:
		configMapName := mountedRuntimeName + "-config"
		err := copyConfigMap(client, configMapName, mountedRuntimeNamespace, configMapName, refNameSpace, ownerReference)
		if err != nil {
			return err
		}
	case common.GooseFSRuntime:
		configMapName := mountedRuntimeName + "-config"
		err := copyConfigMap(client, configMapName, mountedRuntimeNamespace, configMapName, refNameSpace, ownerReference)
		if err != nil {
			return err
		}
	case common.JindoRuntime:
		clientConfigMapName := mountedRuntimeName + "-jindofs-client-config"
		err := copyConfigMap(client, clientConfigMapName, mountedRuntimeNamespace, clientConfigMapName, refNameSpace, ownerReference)
		if err != nil {
			return err
		}
		configMapName := mountedRuntimeName + "-jindofs-config"
		err = copyConfigMap(client, configMapName, mountedRuntimeNamespace, configMapName, refNameSpace, ownerReference)
		if err != nil {
			return err
		}
	case common.ThinRuntime:
		runtimesetConfigMapName := mountedRuntimeName + "-runtimeset"
		err := copyConfigMap(client, runtimesetConfigMapName, mountedRuntimeNamespace, runtimesetConfigMapName, refNameSpace, ownerReference)
		if err != nil {
			return err
		}
	default:
		err := fmt.Errorf("fail to get configmap for runtime type: %s", mountedRuntimeType)
		return err
	}

	return nil
}

func copyConfigMap(client client.Client, srcName string, srcNameSpace string, dstName string, dstNameSpace string, reference metav1.OwnerReference) error {
	found, err := kubeclient.IsConfigMapExist(client, srcName, dstNameSpace)
	if err != nil {
		return err
	}
	if found {
		return nil
	}

	// copy configmap
	srcConfigMap, err := kubeclient.GetConfigmapByName(client, srcName, srcNameSpace)
	if err != nil {
		return err
	}
	// if the source dataset configmap not found, return error and requeue
	if srcConfigMap == nil {
		return fmt.Errorf("runtime configmap %s do not exist", srcName)
	}
	// create the virtual dataset configmap if not exist
	copiedConfigMap := srcConfigMap.DeepCopy()

	dstConfigMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            dstName,
			Namespace:       dstNameSpace,
			Labels:          copiedConfigMap.Labels,
			Annotations:     copiedConfigMap.Annotations,
			OwnerReferences: []metav1.OwnerReference{reference},
		},
		Data: copiedConfigMap.Data,
	}

	err = client.Create(context.TODO(), dstConfigMap)
	if err != nil {
		if otherErr := utils.IgnoreAlreadyExists(err); otherErr != nil {
			return err
		}
	}
	return nil
}
