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
	"github.com/fluid-cloudnative/fluid/pkg/csi/mountinfo"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubelet"
	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/mount"
	"time"
)

type Manager struct {
	mount.SafeFormatAndMount
	KubeletClient *kubelet.KubeletClient
	Driver        *PodDriver

	containers map[string]*containerStat
}

type containerStat struct {
	name    string
	podName string
	startAt metav1.Time
}

func (m *Manager) Run(period int, stopCh <-chan struct{}) {
	go wait.Until(m.run, time.Duration(period)*time.Second, stopCh)
	<-stopCh
	glog.V(3).Info("Shutdown CSI manager.")
}

func (m *Manager) run() {
	pods, err := m.KubeletClient.GetNodeRunningPods()
	glog.V(6).Info("get pods from kubelet")
	if err != nil {
		glog.Error(err)
		return
	}
	for i := range pods.Items {
		pod := pods.Items[i]
		glog.V(6).Infof("get pod: %v", pod)
		if !utils.IsFusePod(pod) {
			continue
		}
		if err := m.Driver.run(&pod); err != nil {
			glog.Error(err)
			return
		}
	}

	brokenMounts, err := mountinfo.GetBrokenMountPoints()
	if err != nil {
		glog.Error(err)
		return
	}
	glog.V(4).Infof("Get broken mount point: %v", brokenMounts)

	go func() {
		for _, point := range brokenMounts {
			m.recoverBrokenMount(point)
		}
	}()
}

func (m *Manager) recoverBrokenMount(point mountinfo.MountPoint) {
	glog.V(3).Infof("Start recovery: [%s], source path: [%s]", point.MountPath, point.SourcePath)
	// recovery for each bind mount path
	mountOption := []string{"bind"}
	if point.ReadOnly {
		mountOption = append(mountOption, "ro")
	}

	glog.V(3).Infof("Start exec cmd: mount %s %s -o %v \n", point.SourcePath, point.MountPath, mountOption)
	if err := m.Mount(point.SourcePath, point.MountPath, "none", mountOption); err != nil {
		glog.Errorf("exec cmd: mount -o bind %s %s err:%v", point.SourcePath, point.MountPath, err)
	}
}

func (m *Manager) compareOrRecordContainerStat(pod corev1.Pod) (restarted bool) {
	if pod.Status.ContainerStatuses == nil {
		return
	}
	for _, cn := range pod.Status.ContainerStatuses {
		if cn.State.Running == nil {
			continue
		}
		cs, ok := m.containers[cn.Name]
		if !ok {
			cs = &containerStat{
				name:    cn.Name,
				podName: pod.Name,
				startAt: cn.State.Running.StartedAt,
			}
			m.containers[cn.Name] = cs
			continue
		}

		if cs.startAt.Before(&cn.State.Running.StartedAt) {
			restarted = true
			return
		}
	}
	return
}
