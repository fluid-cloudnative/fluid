/*

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

package csi

import (
	"os"
	"os/exec"
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/utils/mount"
)

type nodeServer struct {
	nodeId string
	*csicommon.DefaultNodeServer
}

func (ns *nodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {

	glog.Infof("NodePublishVolumeRequest is %v", req)
	targetPath := req.GetTargetPath()

	notMnt, err := mount.New("").IsLikelyNotMountPoint(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(targetPath, 0750); err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			} else {
				glog.Infof("MkdirAll successful. %v", targetPath)
			}
			notMnt = true
		} else {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	if !notMnt {
		glog.Infof("It's already mounted to %v", targetPath)
		return &csi.NodePublishVolumeResponse{}, nil
	} else {
		glog.Infof("Try to mount to %v", targetPath)
	}

	// mountOptions := req.GetVolumeCapability().GetMount().GetMountFlags()
	// if req.GetReadonly() {
	// 	mountOptions = append(mountOptions, "ro")
	// }

	/*
	   https://docs.alluxio.io/os/user/edge/en/api/POSIX-API.html
	   https://github.com/Alluxio/alluxio/blob/master/integration/fuse/bin/alluxio-fuse
	*/

	fluidPath := req.GetVolumeContext()["fluid_path"]
	if fluidPath == "" {
		// fluidPath = fmt.Sprintf("/mnt/%s", req.)
		return nil, status.Error(codes.InvalidArgument, "fluid_path is not set")
	}

	// 1. Wait the alluxio fuse ready
	err = checkMountReady(fluidPath)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	args := []string{"--bind"}
	// if len(mountOptions) > 0 {
	// 	args = append(args, "-o", strings.Join(mountOptions, ","))
	// }

	args = append(args, fluidPath, targetPath)
	command := exec.Command("mount", args...)

	glog.V(4).Infoln(command)
	stdoutStderr, err := command.CombinedOutput()
	glog.V(4).Infoln(string(stdoutStderr))
	if err != nil {
		if os.IsPermission(err) {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		if strings.Contains(err.Error(), "invalid argument") {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &csi.NodePublishVolumeResponse{}, nil
}

func (ns *nodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	targetPath := req.GetTargetPath()

	command := exec.Command("umount",
		targetPath,
	)
	stdoutStderr, err := command.CombinedOutput()
	if err != nil {
		glog.V(3).Infoln(err)
	}
	glog.V(4).Infoln(string(stdoutStderr))

	err = mount.CleanupMountPoint(req.GetTargetPath(), mount.New(""), false)
	if err != nil {
		glog.V(3).Infoln(err)
	}

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (ns *nodeServer) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	return &csi.NodeUnstageVolumeResponse{}, nil
}

func (ns *nodeServer) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	return &csi.NodeStageVolumeResponse{}, nil
}

func (ns *nodeServer) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
