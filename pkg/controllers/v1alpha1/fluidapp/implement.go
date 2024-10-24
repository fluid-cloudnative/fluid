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

package fluidapp

import (
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

type FluidAppReconcilerImplement struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
}

func NewFluidAppReconcilerImplement(client client.Client, log logr.Logger, recorder record.EventRecorder) *FluidAppReconcilerImplement {
	r := &FluidAppReconcilerImplement{
		Client:   client,
		Log:      log,
		Recorder: recorder,
	}
	return r
}

func (i *FluidAppReconcilerImplement) umountFuseSidecars(pod *corev1.Pod) (err error) {
	for _, cn := range pod.Spec.Containers {
		if strings.Contains(cn.Name, common.FuseContainerName) {
			if e := i.umountFuseSidecar(pod, cn); e != nil {
				return
			}
		}
	}
	return
}

func (i *FluidAppReconcilerImplement) umountFuseSidecar(pod *corev1.Pod, fuseContainer corev1.Container) (err error) {
	if fuseContainer.Name == "" {
		return
	}

	cmd := []string{}
	// get prestop
	if fuseContainer.Lifecycle != nil && fuseContainer.Lifecycle.PreStop != nil && fuseContainer.Lifecycle.PreStop.Exec != nil {
		cmd = fuseContainer.Lifecycle.PreStop.Exec.Command
	}

	// if there is no prestop in fuse container, umount mountpath
	if len(cmd) == 0 {
		var mountPath string
		mountPath, err = kubeclient.GetMountPathInContainer(fuseContainer)
		if err != nil {
			return
		}
		if mountPath == "" {
			return nil
		}
		cmd = []string{"umount", mountPath}
	}

	i.Log.Info("exec cmd in pod fuse container", "cmd", cmd, "podName", pod.Name, "namespace", pod.Namespace)
	stdout, stderr, err := kubeclient.ExecCommandInContainer(pod.Name, fuseContainer.Name, pod.Namespace, cmd)
	if err != nil {
		i.Log.Info("exec output", "stdout", stdout, "stderr", stderr)
		if strings.Contains(stderr, "not mounted") {
			// if mount point not mounted, do not retry
			return nil
		}
		if strings.Contains(err.Error(), "exit code 137") {
			i.Log.Info("exec with exit code 137, ignore it.", "err", err)
			return nil
		}
		return err
	}
	return err
}
