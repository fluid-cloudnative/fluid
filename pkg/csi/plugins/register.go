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

package plugins

import (
	"os"

	"github.com/fluid-cloudnative/fluid/pkg/csi/config"
	"github.com/fluid-cloudnative/fluid/pkg/utils/compatibility"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubelet"
	"github.com/golang/glog"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// isUseKubeletConfig checks if the specified Kubelet configuration file exists. If the file does not exist, it returns an empty client without an error .
// If the file exists, it attempts to initialize and return a node-authorized Kubernetes client.
func isUseKubeletConfig(kubeletConfigPath string) bool {
	_, err := os.Stat(kubeletConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			glog.Warningf("kubelet config file %s not exists, continue without node authorization...", kubeletConfigPath)
			return false
		}
		glog.Warningf("fail to stat kubelet config file %s", kubeletConfigPath)
	}

	return true
}

// Register initializes the csi driver and registers it to the controller manager.
func Register(mgr manager.Manager, ctx config.RunningContext) error {
	nodeAuthClient, err := getNodeAuthClient(mgr, ctx)
	if err != nil {
		return err
	}

	csiDriver := NewDriver(ctx.NodeId, ctx.Endpoint, mgr.GetClient(), mgr.GetAPIReader(), nodeAuthClient, ctx.VolumeLocks)

	if err := mgr.Add(csiDriver); err != nil {
		return err
	}

	return nil
}

func getNodeAuthClient(mgr manager.Manager, ctx config.RunningContext) (NodeAuthorizedClient, error) {
	// use and support node binding token
	if !isUseKubeletConfig(ctx.KubeletConfigPath) && compatibility.IsNodeBindingTokenSupported() {
		return &restrictedNodeClient{mgr.GetClient()}, nil
	}
	// otherwise, use kubelet config
	nodeAuthClient, err := kubelet.InitNodeAuthorizedClient(ctx.KubeletConfigPath)
	if err != nil {
		return nil, err
	}
	return &kubeletNodeClient{nodeAuthClient}, nil
}

// Enabled checks if the csi driver should be enabled.
func Enabled() bool {
	return true
}
