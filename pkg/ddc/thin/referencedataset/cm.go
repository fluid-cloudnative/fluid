/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
*/

package referencedataset

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

func copyFuseDaemonSetForRefDataset(client client.Client, refDataset *datav1alpha1.Dataset, physicalRuntimeInfo base.RuntimeInfoInterface) error {
	var fuseName string
	switch physicalRuntimeInfo.GetRuntimeType() {
	case common.JindoRuntime:
		fuseName = physicalRuntimeInfo.GetName() + "-" + common.JindoChartName + "-fuse"
	default:
		fuseName = physicalRuntimeInfo.GetName() + "-fuse"
	}
	ds, err := kubeclient.GetDaemonset(client, fuseName, physicalRuntimeInfo.GetNamespace())
	if err != nil {
		return err
	}

	// copy fuse daemonset to refDataset's namespace
	ownerReference := metav1.OwnerReference{
		APIVersion: refDataset.APIVersion,
		Kind:       refDataset.Kind,
		Name:       refDataset.Name,
		UID:        refDataset.UID,
	}

	dsToCreate := &appsv1.DaemonSet{}
	dsToCreate.Name = refDataset.Name + "-fuse"
	dsToCreate.Namespace = refDataset.Namespace
	dsToCreate.OwnerReferences = append(dsToCreate.OwnerReferences, ownerReference)
	dsToCreate.Spec = *ds.Spec.DeepCopy()
	if len(dsToCreate.Spec.Template.Spec.NodeSelector) == 0 {
		dsToCreate.Spec.Template.Spec.NodeSelector = map[string]string{}
	}
	dsToCreate.Spec.Template.Spec.NodeSelector["fluid.io/fuse-balloon"] = "true"

	err = client.Create(context.TODO(), dsToCreate)
	if utils.IgnoreAlreadyExists(err) != nil {
		return err
	}

	return nil
}

func (e *ReferenceDatasetEngine) createConfigMapForRefDataset(client client.Client, refDataset *datav1alpha1.Dataset, physicalRuntimeInfo base.RuntimeInfoInterface) error {
	physicalRuntimeType := physicalRuntimeInfo.GetRuntimeType()
	physicalRuntimeName := physicalRuntimeInfo.GetName()
	physicalRuntimeNamespace := physicalRuntimeInfo.GetNamespace()

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
	switch physicalRuntimeType {
	// TODO:  currently the dst configmap name is the same as src configmap name to avoid modify the fuse init container filed,
	//       but duplicated name error can occurs if the dst namespace has same named runtime.
	case common.AlluxioRuntime:
		configMapName := physicalRuntimeName + "-config"
		err := kubeclient.CopyConfigMap(client, types.NamespacedName{Name: configMapName, Namespace: physicalRuntimeNamespace},
			types.NamespacedName{Name: configMapName, Namespace: refNameSpace}, ownerReference)
		if err != nil {
			return err
		}
	case common.JuiceFSRuntime:
		fuseScriptConfigMapName := physicalRuntimeName + "-fuse-script"
		err := kubeclient.CopyConfigMap(client, types.NamespacedName{Name: fuseScriptConfigMapName, Namespace: physicalRuntimeNamespace},
			types.NamespacedName{Name: fuseScriptConfigMapName, Namespace: refNameSpace}, ownerReference)
		if err != nil {
			return err
		}
	case common.GooseFSRuntime:
		configMapName := physicalRuntimeName + "-config"
		err := kubeclient.CopyConfigMap(client, types.NamespacedName{Name: configMapName, Namespace: physicalRuntimeNamespace},
			types.NamespacedName{Name: configMapName, Namespace: refNameSpace}, ownerReference)
		if err != nil {
			return err
		}
	case common.JindoRuntime:
		clientConfigMapName := physicalRuntimeName + "-jindofs-client-config"
		err := kubeclient.CopyConfigMap(client, types.NamespacedName{Name: clientConfigMapName, Namespace: physicalRuntimeNamespace},
			types.NamespacedName{Name: clientConfigMapName, Namespace: refNameSpace}, ownerReference)
		if err != nil {
			return err
		}
		configMapName := physicalRuntimeName + "-jindofs-config"
		err = kubeclient.CopyConfigMap(client, types.NamespacedName{Name: configMapName, Namespace: physicalRuntimeNamespace},
			types.NamespacedName{Name: configMapName, Namespace: refNameSpace}, ownerReference)
		if err != nil {
			return err
		}
	case common.EFCRuntime:
		// TODO: EFCRuntime needs worker-endpoint configmap which should be synced timely for ECI mode.
		// Currently EFCRuntime only supports CSI mode, so do nothing here.
		e.Log.Info("Skip createConfigMapForRefDataset because the physicalRuntimeType=EFC", "name", e.name, "namespace", e.namespace)
	case common.ThinRuntime:
		runtimesetConfigMapName := physicalRuntimeName + "-runtimeset"
		err := kubeclient.CopyConfigMap(client, types.NamespacedName{Name: runtimesetConfigMapName, Namespace: physicalRuntimeNamespace},
			types.NamespacedName{Name: runtimesetConfigMapName, Namespace: refNameSpace}, ownerReference)
		if err != nil {
			return err
		}
	default:
		err := fmt.Errorf("fail to get configmap for runtime type: %s", physicalRuntimeType)
		return err
	}

	return nil
}
