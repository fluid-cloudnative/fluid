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
	"context"
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/csi/util"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	k8sexec "k8s.io/utils/exec"
	"k8s.io/utils/mount"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PodDriver struct {
	client.Client
	handlers map[podStatus]podHandler
	mount.SafeFormatAndMount
}

func NewPodDriver(client client.Client) *PodDriver {
	mounter := &mount.SafeFormatAndMount{
		Interface: mount.New(""),
		Exec:      k8sexec.New(),
	}
	driver := &PodDriver{
		Client:             client,
		handlers:           map[podStatus]podHandler{},
		SafeFormatAndMount: *mounter,
	}
	driver.handlers[podReady] = driver.podReadyHandler
	driver.handlers[podError] = driver.podErrorHandler
	driver.handlers[podPending] = driver.podPendingHandler
	driver.handlers[podDeleted] = driver.podDeletedHandler
	return driver
}

type podHandler func(ctx context.Context, pod *corev1.Pod) error
type podStatus string

const (
	podReady   podStatus = "podReady"
	podError   podStatus = "podError"
	podDeleted podStatus = "podDeleted"
	podPending podStatus = "podPending"
)

func (p *PodDriver) run(ctx context.Context, pod *corev1.Pod) error {
	return p.handlers[p.getPodStatus(pod)](ctx, pod)
}

func (p *PodDriver) getPodStatus(pod *corev1.Pod) podStatus {
	if pod == nil {
		return podError
	}
	if pod.DeletionTimestamp != nil {
		return podDeleted
	}
	if utils.IsPodError(pod) {
		return podError
	}
	if utils.IsPodReady(pod) {
		return podReady
	}
	return podPending
}

func (p *PodDriver) podReadyHandler(ctx context.Context, pod *corev1.Pod) (err error) {
	if pod == nil {
		klog.Errorf("[podReadyHandler] get nil pod")
		return nil
	}

	// get runtime name
	runtimeName, err := util.GetRuntimeNameFromFusePod(*pod)
	if err != nil {
		return
	}
	// get pv
	pvName := fmt.Sprintf("%s-%s", pod.Namespace, runtimeName)
	pv, err := kubeclient.GetPersistentVolume(p.Client, pvName)

	// get mount point
	mountPath := pv.Spec.CSI.VolumeAttributes[common.FLUID_PATH]

	// get target path
	targets, err := util.GetPVMountPoint(pvName)
	if err != nil {
		return
	}

	readOnly := false
	// get access mode
	// see: https://github.com/kubernetes/kubernetes/blob/master/pkg/volume/csi/csi_mounter.go#L171
	accessMode := corev1.ReadWriteOnce
	if pv.Spec.AccessModes != nil {
		accessMode = pv.Spec.AccessModes[0]
	}
	if accessMode == corev1.ReadOnlyMany {
		readOnly = true
		glog.Infof("Set the mount option readonly=%v", readOnly)
	}

	// recovery for each target
	mountOption := []string{"bind"}
	if readOnly {
		mountOption = append(mountOption, "ro")
	}
	for i := range targets {
		target := targets[i]
		// check target should do recover
		corruptedMnt, err := util.CheckMountPointBroken(target)
		if err != nil {
			klog.V(5).Infof("target path %s is normal, don't need do recover", target)
			continue
		}
		if !corruptedMnt {
			klog.V(5).Infof("target %s not exists,  don't do recover", target)
			continue
		}
		klog.Infof("start exec cmd: mount -o bind %s %s \n", mountPath, target)
		if err := p.Mount(mountPath, target, "none", mountOption); err != nil {
			klog.Errorf("exec cmd: mount -o bind %s %s err:%v", mountPath, target, err)
		}
	}
	return nil
}

func (p *PodDriver) podPendingHandler(ctx context.Context, pod *corev1.Pod) error {
	return nil
}

func (p *PodDriver) podErrorHandler(ctx context.Context, pod *corev1.Pod) error {
	return nil
}

func (p *PodDriver) podDeletedHandler(ctx context.Context, pod *corev1.Pod) error {
	return nil
}
