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
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/wait"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/utils/kubelet"
)

const DefaultCSIManagerPeriod = 10

type Manager struct {
	KubeletClient *kubelet.KubeletClient
	Driver        *PodDriver
}

func (m *Manager) Run(ctx context.Context) {
	wait.Forever(func() {
		m.run(ctx)
	}, DefaultCSIManagerPeriod*time.Second)
}

func (m *Manager) run(ctx context.Context) {
	pods, err := m.KubeletClient.GetNodeRunningPods()
	if err != nil {
		glog.Error(err)
		return
	}
	for i := range pods.Items {
		pod := pods.Items[i]
		// todo query pod
		if err := m.Driver.run(ctx, &pod); err != nil {
			glog.Error(err)
			return
		}
	}
}
