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

	locks *utils.VolumeLocks
}

var _ manager.Runnable = &driver{}

func NewDriver(nodeID, endpoint string, client client.Client, apiReader client.Reader, nodeAuthorizedClient *kubernetes.Clientset, locks *utils.VolumeLocks) *driver {
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
		locks:                locks,
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
		locks:                d.locks,
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
