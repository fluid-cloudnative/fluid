/*
Copyright 2023 The Fluid Authors.

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
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/golang/glog"
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
)

const (
	driverName = "fuse.csi.fluid.io"
	version    = "1.0.0"
)

type driver struct {
	client               client.Client
	apiReader            client.Reader
	nodeAuthorizedClient *kubernetes.Clientset
	csiDriver            *csicommon.CSIDriver
	nodeId, endpoint     string
}

var _ manager.Runnable = &driver{}

func NewDriver(nodeID, endpoint string, client client.Client, apiReader client.Reader, nodeAuthorizedClient *kubernetes.Clientset) *driver {
	glog.Infof("Driver: %v version: %v", driverName, version)

	proto, addr := utils.SplitSchemaAddr(endpoint)
	glog.Infof("protocol: %v addr: %v", proto, addr)

	if !strings.HasPrefix(addr, "/") {
		addr = fmt.Sprintf("/%s", addr)
	}

	socketDir := filepath.Dir(addr)
	err := os.MkdirAll(socketDir, 0755)
	if err != nil {
		glog.Errorf("failed due to %v", err)
		os.Exit(1)
	}

	csiDriver := csicommon.NewCSIDriver(driverName, version, nodeID)
	csiDriver.AddControllerServiceCapabilities([]csi.ControllerServiceCapability_RPC_Type{csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME})
	csiDriver.AddVolumeCapabilityAccessModes([]csi.VolumeCapability_AccessMode_Mode{csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER})

	return &driver{
		nodeId:               nodeID,
		endpoint:             endpoint,
		csiDriver:            csiDriver,
		client:               client,
		nodeAuthorizedClient: nodeAuthorizedClient,
		apiReader:            apiReader,
	}
}

func (d *driver) newControllerServer() *controllerServer {
	return &controllerServer{
		DefaultControllerServer: csicommon.NewDefaultControllerServer(d.csiDriver),
	}
}

func (d *driver) newNodeServer() *nodeServer {
	return &nodeServer{
		nodeId:               d.nodeId,
		DefaultNodeServer:    csicommon.NewDefaultNodeServer(d.csiDriver),
		client:               d.client,
		apiReader:            d.apiReader,
		nodeAuthorizedClient: d.nodeAuthorizedClient,
	}
}

func (d *driver) run() {
	s := csicommon.NewNonBlockingGRPCServer()
	s.Start(
		d.endpoint,
		csicommon.NewDefaultIdentityServer(d.csiDriver),
		d.newControllerServer(),
		d.newNodeServer(),
	)
	s.Wait()
}

func (d *driver) Start(ctx context.Context) error {
	d.run()
	return nil
}
