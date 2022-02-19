/*
Copyright 2021 The Fluid Authors.

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

package volume

import (
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetNamespacedNameByVolumeId(client client.Client, volumeId string) (namespace, name string, err error) {
	pv, err := kubeclient.GetPersistentVolume(client, volumeId)
	if err != nil {
		return "", "", err
	}

	if pv.Spec.ClaimRef == nil {
		return "", "", errors.Errorf("pv %s has unexpected nil claimRef", volumeId)
	}

	namespace = pv.Spec.ClaimRef.Namespace
	name = pv.Spec.ClaimRef.Name

	ok, err := kubeclient.IsDatasetPVC(client, name, namespace)
	if err != nil {
		return "", "", err
	}

	if !ok {
		return "", "", errors.Errorf("pv %s is not bounded with a fluid pvc", volumeId)
	}

	return
}
