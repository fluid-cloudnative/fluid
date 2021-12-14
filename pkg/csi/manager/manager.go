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
	"github.com/fluid-cloudnative/fluid/pkg/csi/util"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/wait"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/utils/kubelet"
)

type Manager struct {
	KubeletClient *kubelet.KubeletClient
	Driver        *PodDriver
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
	go func() {
		for i := range pods.Items {
			pod := pods.Items[i]
			glog.V(6).Infof("get pod: %v", pod)
			if !util.IsFusePod(pod) {
				continue
			}
			if err := m.Driver.run(&pod); err != nil {
				glog.Error(err)
				return
			}
		}
	}()
}
