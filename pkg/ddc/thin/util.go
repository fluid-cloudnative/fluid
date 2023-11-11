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

func (t *ThinEngine) getFuseDaemonsetName() (dsName string) {
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
