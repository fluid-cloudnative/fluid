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

package manager

import (
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
	k8sexec "k8s.io/utils/exec"
	"k8s.io/utils/mount"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PodDriver struct {
	client.Client
	mount.SafeFormatAndMount
}

func NewPodDriver(client client.Client) *PodDriver {
	return &PodDriver{
		Client: client,
		SafeFormatAndMount: mount.SafeFormatAndMount{
			Interface: mount.New(""),
			Exec:      k8sexec.New(),
		},
	}
}

func (p *PodDriver) run(pod *corev1.Pod) error {
	if podutil.IsPodReady(pod) {
		// Only handle when pod ready
		return p.podReadyHandler(pod)
	}
	return nil
}

func (p *PodDriver) podReadyHandler(pod *corev1.Pod) (err error) {
	if pod == nil {
		glog.Error("[podReadyHandler] get nil pod")
		return nil
	}
	glog.V(3).Infof("Get ready pod: [%s-%s]", pod.Namespace, pod.Name)

	// get runtime name
	runtimeName, err := utils.GetRuntimeNameFromFusePod(*pod)
	if err != nil {
		return
	}
	glog.V(3).Infof("Get runtimeName %s from pod", runtimeName)
	// get pv
	pvName := fmt.Sprintf("%s-%s", pod.Namespace, runtimeName)
	glog.V(3).Infof("Get pvName %s", pvName)
	pv, err := kubeclient.GetPersistentVolume(p.Client, pvName)

	// get mount point
	mountPath := pv.Spec.CSI.VolumeAttributes[common.FLUID_PATH]
	glog.V(3).Infof("Get mountPath %s", mountPath)

	return nil
}
