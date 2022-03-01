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
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

func (i *FluidAppReconcilerImplement) shouldReconcile(ctx reconcileRequestContext) bool {
	pod := ctx.pod
	if pod == nil {
		return false
	}

	// ignore if restartPolicy is Always
	if pod.Spec.RestartPolicy == corev1.RestartPolicyAlways {
		return false
	}

	// ignore if no fuse container
	exist := false
	for _, cn := range pod.Spec.Containers {
		if cn.Name == common.FuseContainerName {
			exist = true
		}
	}
	if !exist {
		i.Log.Info("There are no fluid fuse sidecar in pod.", "name", pod.Name, "namespace", pod.Namespace)
		return false
	}

	// reconcile if app container exit 0 and fuse container not exit
	antiFuseExited := false
	fuseExited := false
	for _, containerStatus := range pod.Status.ContainerStatuses {
		//i.Log.Info("container status", "name", containerStatus.Name, "status", containerStatus.State)
		if containerStatus.Name != common.FuseContainerName {
			if containerStatus.State.Terminated != nil && containerStatus.State.Terminated.ExitCode == 0 {
				antiFuseExited = true
			}
		}
		if containerStatus.Name == common.FuseContainerName {
			if containerStatus.State.Terminated != nil && containerStatus.State.Terminated.ExitCode == 0 {
				fuseExited = true
			}
		}
	}
	return antiFuseExited == true && fuseExited == false
}

func (i *FluidAppReconcilerImplement) umountFuseSidecar(pod *corev1.Pod) (err error) {
	var fuseContainer corev1.Container
	for _, cn := range pod.Spec.Containers {
		if cn.Name == common.FuseContainerName {
			fuseContainer = cn
		}
	}
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
	stdout, stderr, err := kubeclient.ExecCommandInContainer(pod.Name, common.FuseContainerName, pod.Namespace, cmd)
	if err != nil {
		i.Log.Info("exec output", "stdout", stdout, "stderr", stderr)
		return err
	}
	return err
}
