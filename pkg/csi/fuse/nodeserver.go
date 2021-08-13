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
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/pkg/errors"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog"
	"os"
	"os/exec"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"syscall"

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
	client client.Client
}

func (ns *nodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {

	glog.Infof("NodePublishVolumeRequest is %v", req)
	targetPath := req.GetTargetPath()

	isMount, err := isMounted(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(targetPath, 0750); err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			} else {
				glog.Infof("MkdirAll successful. %v", targetPath)
			}
			//isMount = true
		} else {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	if isMount {
		glog.Infof("It's already mounted to %v", targetPath)
		return &csi.NodePublishVolumeResponse{}, nil
	} else {
		glog.Infof("Try to mount to %v", targetPath)
	}

	// 0. check if read only
	readOnly := false
	if req.GetVolumeCapability() == nil {
		glog.Infoln("Volume Capability is nil")
	} else {
		mode := req.GetVolumeCapability().GetAccessMode().GetMode()
		if mode == csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY {
			readOnly = true
			glog.Infof("Set the mount option readonly=%v", readOnly)
		}
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
	mountType := req.GetVolumeContext()["mount_type"]
	if fluidPath == "" {
		// fluidPath = fmt.Sprintf("/mnt/%s", req.)
		return nil, status.Error(codes.InvalidArgument, "fluid_path is not set")
	}
	if mountType == "" {
		// default mountType is ALLUXIO_MOUNT_TYPE
		mountType = common.ALLUXIO_MOUNT_TYPE
	}

	// 1. Wait the runtime fuse ready
	err = checkMountReady(fluidPath, mountType)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	args := []string{"--bind"}
	// if len(mountOptions) > 0 {
	// 	args = append(args, "-o", strings.Join(mountOptions, ","))
	// }

	if readOnly {
		args = append(args, "-o", "ro", fluidPath, targetPath)
	} else {
		args = append(args, fluidPath, targetPath)
	}
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
	} else {
		glog.V(4).Infof("Succeed in binding %s to %s", fluidPath, targetPath)
	}

	return &csi.NodePublishVolumeResponse{}, nil
}

func (ns *nodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	targetPath := req.GetTargetPath()

	command := exec.Command("umount",
		targetPath,
	)
	glog.V(4).Infoln(command)
	stdoutStderr, err := command.CombinedOutput()
	if err != nil {
		glog.V(3).Infoln(err)
	}
	glog.V(4).Infoln(string(stdoutStderr))

	err = mount.CleanupMountPoint(req.GetTargetPath(), mount.New(""), false)
	if err != nil {
		glog.V(3).Infoln(err)
	} else {
		glog.V(4).Infof("Succeed in umounting  %s", targetPath)
	}

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (ns *nodeServer) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	// 1. get fuse label according to pv namespace and name
	namespacedName := strings.Split(req.GetVolumeId(), "-")
	fuseLabelKey := common.LabelAnnotationFusePrefix + namespacedName[0] + "-" + namespacedName[1]

	// 2. check if the path is mounted
	inUse, err := checkMountInUse(req.GetVolumeId())
	if err != nil {
		return nil, errors.Wrap(err, "NodeUnstageVolume: can't check mount in use")
	}
	if inUse {
		return nil, fmt.Errorf("NodeUnstageVolume: can't stop fuse cause it's in use")
	}

	// 3. remove label on node.
	// Once the label is removed, fuse pod on corresponding node will be terminated
	// since node selector in the fuse daemonSet no longer matches.
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		node, err := kubeclient.GetNode(ns.client, ns.nodeId)
		if err != nil {
			return err
		}

		nodeToUpdate := node.DeepCopy()
		delete(nodeToUpdate.Labels, fuseLabelKey)
		if !reflect.DeepEqual(node, nodeToUpdate) {
			return ns.client.Update(context.TODO(), nodeToUpdate)
		} else {
			klog.Infof("Do nothing because no label %s on node %s", fuseLabelKey, ns.nodeId)
		}

		return nil
	})

	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("NodeUnstageVolume: can't remove label %s on node %s", fuseLabelKey, ns.nodeId))
	}
	return &csi.NodeUnstageVolumeResponse{}, nil
}

func (ns *nodeServer) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	// 1. get fuse label according to pv namespace and name
	namespacedName := strings.Split(req.GetVolumeId(), "-")
	fuseLabelKey := common.LabelAnnotationFusePrefix + namespacedName[0] + "-" + namespacedName[1]

	// 2. Label node
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		node, err := kubeclient.GetNode(ns.client, ns.nodeId)
		if err != nil {
			return err
		}

		nodeToLabel := node.DeepCopy()
		nodeToLabel.Labels[fuseLabelKey] = "true"
		if !reflect.DeepEqual(node, nodeToLabel) {
			return ns.client.Update(context.TODO(), nodeToLabel)
		} else {
			klog.Infof("Do nothing because label %s already added on node %s", fuseLabelKey, ns.nodeId)
		}

		return nil
	})

	if err != nil {
		klog.Infof("NodeStageVolume: can't add label %s on node %s", fuseLabelKey, ns.nodeId)
		return nil, errors.Wrap(err, fmt.Sprintf("NodeStageVolume: can't add label %s on node %s", fuseLabelKey, ns.nodeId))
	}

	fluidPath := req.GetVolumeContext()["fluid_path"]
	mountType := req.GetVolumeContext()["mount_type"]

	// checkMountReady checks the fuse mount path every 3 second for 30 seconds in total.
	err = checkMountReady(fluidPath, mountType)
	if err != nil {
		return nil, errors.Errorf("fuse pod on node %s is not ready", ns.nodeId)
	}

	return &csi.NodeStageVolumeResponse{}, nil
}

func (ns *nodeServer) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (ns *nodeServer) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	glog.V(5).Infof("Using default NodeGetCapabilities")

	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: []*csi.NodeServiceCapability{
			{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
					},
				},
			},
		},
	}, nil
}

func checkMountInUse(volumeName string) (bool, error) {
	var inUse bool
	glog.Infof("Try to check if the volume %s is being used", volumeName)
	if volumeName == "" {
		return inUse, errors.New("volumeName is not specified")
	}

	// TODO: refer to https://github.com/kubernetes-sigs/alibaba-cloud-csi-driver/blob/4fcb743220371de82d556ab0b67b08440b04a218/pkg/oss/utils.go#L72
	// for a better implementation
	command := exec.Command("/usr/local/bin/check_bind_mounts.sh", volumeName)

	if err := command.Start(); err != nil {
		return inUse, err
	}

	err := command.Wait()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				exitStatus := status.ExitStatus()
				if exitStatus == 1 {
					err = nil
					inUse = false
				}
			}
		}
	} else {
		waitStatus := command.ProcessState.Sys().(syscall.WaitStatus)
		if waitStatus.ExitStatus() == 0 {
			inUse = true
		}
		return inUse, fmt.Errorf("unexpcted return code happen when checking mount in use")
	}

	return inUse, err
}
