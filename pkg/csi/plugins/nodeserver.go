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
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/cmdguard"
	"github.com/fluid-cloudnative/fluid/pkg/utils/dataset/volume"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/mount"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	AllowPatchStaleNodeEnv = "ALLOW_PATCH_STALE_NODE"
)

type nodeServer struct {
	nodeId string
	*csicommon.DefaultNodeServer
	client               client.Client
	apiReader            client.Reader
	nodeAuthorizedClient *kubernetes.Clientset
	locks                *utils.VolumeLocks
	node                 *v1.Node
}

func (ns *nodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {

	glog.Infof("NodePublishVolumeRequest is %v", req)
	targetPath := req.GetTargetPath()
	// check targetpath validity
	if len(targetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "NodePublishVolume operation requires targetPath but is not provided")
	}

	// The lock is to avoid race condition
	if lock := ns.locks.TryAcquire(targetPath); !lock {
		return nil, status.Errorf(codes.Aborted, "NodePublishVolume operation on targetPath %s already exists", targetPath)
	}
	defer ns.locks.Release(targetPath)

	isMount, err := utils.IsMounted(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(targetPath, 0750); err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			} else {
				glog.Infof("NodePublishVolume: MkdirAll successful on %v", targetPath)
			}
			//isMount = true
		} else {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	if isMount {
		glog.Infof("NodePublishVolume: already mounted to %v, do nothing", targetPath)
		return &csi.NodePublishVolumeResponse{}, nil
	}

	glog.Infof("NodePublishVolume: start mounting staging path to %v", targetPath)
	// 0. check if read only
	readOnly := false
	if req.GetVolumeCapability() == nil {
		glog.Infoln("NodePublishVolume: found volume capability is nil")
	} else {
		mode := req.GetVolumeCapability().GetAccessMode().GetMode()
		if mode == csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY {
			readOnly = true
			glog.Infof("NodePublishVolume: set the mount option readonly=%v", readOnly)
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

	fluidPath := req.GetVolumeContext()[common.VolumeAttrFluidPath]
	mountType := req.GetVolumeContext()[common.VolumeAttrMountType]
	subPath := req.GetVolumeContext()[common.VolumeAttrFluidSubPath]

	if fluidPath == "" {
		// fluidPath = fmt.Sprintf("/mnt/%s", req.)
		return nil, status.Error(codes.InvalidArgument, "fluid_path is not set")
	}
	if mountType == "" {
		// default mountType is ALLUXIO_MOUNT_TYPE
		mountType = common.AlluxioMountType
	}

	mountPath := fluidPath
	if subPath != "" {
		mountPath = fluidPath + "/" + subPath
	}

	// 1. Wait the runtime fuse ready and check the sub path existence
	useSymlink := useSymlink(req)

	skipCheckMountReadyMountModeSelector, err := base.ParseMountModeSelectorFromStr(req.GetVolumeContext()[common.AnnotationSkipCheckMountReadyTarget])
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if skipCheckMountReadyMountModeSelector.Selected(base.MountPodMountMode) {
		// 1. only mountPod involved csi-plugin
		// 2. skip check mount ready for mountPod, for the scenario that dataset.spec.mounts is nil
		// 3. if check mount ready is skipped for mountPod, symlink is forced to use, avoiding that unPublishVolume error occurs
		useSymlink = true
	} else {
		err = utils.CheckMountReadyAndSubPathExist(fluidPath, mountType, subPath)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	// use symlink
	if useSymlink {
		if err := utils.CreateSymlink(targetPath, mountPath); err != nil {
			return nil, err
		}
		return &csi.NodePublishVolumeResponse{}, nil
	}

	// default use bind mount
	args := []string{"--bind"}
	// if len(mountOptions) > 0 {
	// 	args = append(args, "-o", strings.Join(mountOptions, ","))
	// }

	if readOnly {
		args = append(args, "-o", "ro", mountPath, targetPath)
	} else {
		args = append(args, mountPath, targetPath)
	}
	command, err := cmdguard.Command("mount", args...)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	glog.V(3).Infof("NodePublishVolume: exec command %v", command)
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
		glog.V(3).Infof("NodePublishVolume: succeed in binding %s to %s", mountPath, targetPath)
	}

	return &csi.NodePublishVolumeResponse{}, nil
}

// NodeUnpublishVolume umounts every mounted file systems on the given req.GetTargetPath() until it's cleaned up.
// If anything unexpected happened during the umount process, it returns error and wait for retries.
func (ns *nodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	targetPath := req.GetTargetPath()
	// check targetpath validity
	if len(targetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "NodeUnpublishVolume operation requires targetPath but is not provided")
	}

	// The lock is to avoid race condition, make sure only one goroutine(including the FUSE Recovery goroutine) is handling the targetPath
	if lock := ns.locks.TryAcquire(targetPath); !lock {
		return nil, status.Errorf(codes.Aborted, "NodeUnpublishVolume operation on targetPath %s already exists", targetPath)
	}
	defer ns.locks.Release(targetPath)

	exists, err := mount.PathExists(targetPath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "NodeUnpublishVolume: failed to check if path %s exists: %v", targetPath, err)
	}

	if !exists {
		glog.V(0).Infof("NodeUnpublishVolume: succeed because target path %s doesn't exist", targetPath)
		return &csi.NodeUnpublishVolumeResponse{}, nil
	}

	// try to remove if targetPath is a symlink
	symlinkRemove, err := utils.RemoveSymlink(targetPath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "NodeUnpublishVolume: remove symlink error %v", err)
	}
	if symlinkRemove {
		// targetPath is a symlink and has been remove successfully
		glog.V(3).Infof("Remove symlink targetPath %s successfully", targetPath)
		return &csi.NodeUnpublishVolumeResponse{}, nil
	}

	// targetPath may be bind mount many times when mount point recovered.
	// umount until it's not mounted.
	mounter := mount.New("")
	for {
		needUnmount, err := isLikelyNeedUnmount(mounter, targetPath)
		if err != nil {
			glog.Errorf("NodeUnpublishVolume: fail to check if targetPath %s needs unmount: %v", targetPath, err)
			return nil, status.Errorf(codes.Internal, "NodeUnpublishVolume: fail to check if targetPath %s needs unmount: %v", targetPath, err)
		}

		if !needUnmount {
			glog.V(3).Infof("NodeUnpublishVolume: umount %s success", targetPath)
			break
		}

		glog.V(3).Infof("NodeUnpublishVolume: exec umount %s", targetPath)
		err = mounter.Unmount(targetPath)
		if os.IsNotExist(err) {
			glog.V(3).Infof("NodeUnpublishVolume: targetPath %s has been cleaned up when umounting it", targetPath)
			break
		}
		if err != nil {
			glog.Errorf("NodeUnpublishVolume: umount targetPath %s with error: %v", targetPath, err)
			return nil, status.Errorf(codes.Internal, "NodeUnpublishVolume: umount targetPath %s: %v", targetPath, err)
		}
	}

	err = mount.CleanupMountPoint(targetPath, mounter, false)
	if err != nil {
		glog.Errorf("NodeUnpublishVolume: failed when cleanupMountPoint on path %s: %v", targetPath, err)
		return nil, status.Errorf(codes.Internal, "NodeUnpublishVolume: failed when cleanupMountPoint on path %s: %v", targetPath, err)
	} else {
		glog.V(4).Infof("NodeUnpublishVolume: succeed in umounting %s", targetPath)
	}

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (ns *nodeServer) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	volumeId := req.GetVolumeId()
	if len(volumeId) == 0 {
		return nil, status.Error(codes.InvalidArgument, "NodeUnstageVolume operation requires volumeId but is not provided")
	}

	// The lock is to ensure CSI plugin labels the node in correct order
	if lock := ns.locks.TryAcquire(volumeId); !lock {
		return nil, status.Errorf(codes.Aborted, "NodeUnstageVolume operation on volumeId %s already exists", volumeId)
	}
	defer ns.locks.Release(volumeId)

	// 1. get runtime namespace and name
	// A nil volumeContext is passed because unlike csi.NodeStageVolumeRequest, csi.NodeUnstageVolumeRequest has
	// no volume context attribute.
	namespace, name, err := ns.getRuntimeNamespacedName(nil, volumeId)
	if err != nil {
		if utils.IgnoreNotFound(err) == nil {
			// For cases like the related persistent volume has been deleted, ignore it and return success
			glog.Warningf("NodeUnstageVolume: volume %s not found, maybe it's already cleaned up, ignore it", volumeId)
			return &csi.NodeUnstageVolumeResponse{}, nil
		}
		glog.Errorf("NodeUnstageVolume: can't get runtime namespace and name given (volumeContext: nil, volumeId: %s): %v", volumeId, err)
		return nil, errors.Wrapf(err, "NodeUnstageVolume: can't get namespace and name by volume id %s", volumeId)
	}

	// 2. Check fuse clean policy. If clean policy is set to OnRuntimeDeleted, there is no
	// need to clean fuse eagerly.
	runtimeInfo, err := base.GetRuntimeInfo(ns.client, name, namespace)
	if err != nil {
		if utils.IgnoreNotFound(err) == nil {
			// For cases like the dataset or runtime has been deleted, ignore it and return success
			glog.Warningf("NodeUnstageVolume: dataset or runtime %s/%s not found, maybe it's already cleaned up", namespace, name)
			return &csi.NodeUnstageVolumeResponse{}, nil
		}
		return nil, errors.Wrapf(err, "NodeUnstageVolume: failed to get runtime info for %s/%s", namespace, name)
	}

	var shouldCleanFuse bool
	cleanPolicy := runtimeInfo.GetFuseCleanPolicy()
	glog.Infof("NodeUnstageVolume: Using %s clean policy for runtime %s in namespace %s", cleanPolicy, runtimeInfo.GetName(), runtimeInfo.GetNamespace())
	switch cleanPolicy {
	case v1alpha1.OnDemandCleanPolicy:
		shouldCleanFuse = true
	case v1alpha1.OnRuntimeDeletedCleanPolicy:
		shouldCleanFuse = false
	default:
		return nil, errors.Errorf("NodeUnstageVolume: unknown Fuse clean policy: %s", cleanPolicy)
	}

	if !shouldCleanFuse {
		return &csi.NodeUnstageVolumeResponse{}, nil
	}

	// 3. check if the path is mounted
	inUse, err := checkMountInUse(volumeId)
	if err != nil {
		return nil, errors.Wrap(err, "NodeUnstageVolume: can't check mount in use")
	}
	if inUse {
		return nil, fmt.Errorf("NodeUnstageVolume: can't stop fuse cause it's in use")
	}

	// 4. remove label on node
	// Once the label is removed, fuse pod on corresponding node will be terminated
	// since node selector in the fuse daemonSet no longer matches.
	// TODO: move all the label keys into a util func
	fuseLabelKey := common.LabelAnnotationFusePrefix + namespace + "-" + name
	var labelsToModify common.LabelsToModify
	labelsToModify.Delete(fuseLabelKey)

	node, err := ns.getNode()
	if err != nil {
		glog.Errorf("NodeUnstageVolume: can't get node %s: %v", ns.nodeId, err)
		return nil, errors.Wrapf(err, "NodeUnstageVolume: can't get node %s", ns.nodeId)
	}

	// _, err = utils.ChangeNodeLabelWithPatchMode(ns.client, node, labelsToModify)
	err = ns.patchNodeWithLabel(node, labelsToModify)
	if err != nil {
		glog.Errorf("NodeUnstageVolume: error when patching labels on node %s: %v", ns.nodeId, err)
		return nil, errors.Wrapf(err, "NodeUnstageVolume: error when patching labels on node %s", ns.nodeId)
	}

	return &csi.NodeUnstageVolumeResponse{}, nil
}

func (ns *nodeServer) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	volumeId := req.GetVolumeId()
	if len(volumeId) == 0 {
		return nil, status.Error(codes.InvalidArgument, "NodeStageVolume operation requires volumeId but is not provided")
	}

	// The lock is to ensure CSI plugin labels the node in correct order
	if lock := ns.locks.TryAcquire(volumeId); !lock {
		return nil, status.Errorf(codes.Aborted, "NodeStageVolume operation on volumeId %s already exists", volumeId)
	}
	defer ns.locks.Release(volumeId)
	glog.Infof("NodeStageVolume: Starting NodeStage with VolumeId: %s, and VolumeContext: %v", volumeId, req.VolumeContext)

	// 1. Start SessMgr Pod and wait for ready if FUSE pod requires SessMgr
	sessMgrWorkDir := req.GetVolumeContext()[common.VolumeAttrEFCSessMgrWorkDir]
	if len(sessMgrWorkDir) != 0 {
		if err := ns.prepareSessMgr(sessMgrWorkDir); err != nil {
			glog.Errorf("NodeStageVolume: fail to prepare SessMgr because: %v", err)
			return nil, errors.Wrapf(err, "NodeStageVolume: fail to prepare SessMgr")
		}
	}

	// 2. clean up broken mount point
	fluidPath := req.GetVolumeContext()[common.VolumeAttrFluidPath]
	if ignoredErr := cleanUpBrokenMountPoint(fluidPath); ignoredErr != nil {
		glog.Warningf("NodeStageVolume: Ignoring error when cleaning up broken mount point %v: %v", fluidPath, ignoredErr)
	}

	// 3. get runtime namespace and name
	namespace, name, err := ns.getRuntimeNamespacedName(req.GetVolumeContext(), volumeId)
	if err != nil {
		glog.Errorf("NodeStageVolume: can't get runtime namespace and name given (volumeContext: %v, volumeId: %s): %v", req.GetVolumeContext(), req.GetVolumeId(), err)
		return nil, errors.Wrapf(err, "NodeStageVolume: can't get namespace and name by volume id %s", volumeId)
	}

	// 4. Label node to launch FUSE Pod
	fuseLabelKey := common.LabelAnnotationFusePrefix + namespace + "-" + name
	var labelsToModify common.LabelsToModify
	labelsToModify.Add(fuseLabelKey, "true")

	node, err := ns.getNode()
	if err != nil {
		glog.Errorf("NodeStageVolume: can't get node %s: %v", ns.nodeId, err)
		return nil, errors.Wrapf(err, "NodeStageVolume: can't get node %s", ns.nodeId)
	}

	// _, err = utils.ChangeNodeLabelWithPatchMode(ns.client, node, labelsToModify)
	err = ns.patchNodeWithLabel(node, labelsToModify)
	if err != nil {
		glog.Errorf("NodeStageVolume: error when patching labels on node %s: %v", ns.nodeId, err)
		return nil, errors.Wrapf(err, "NodeStageVolume: error when patching labels on node %s", ns.nodeId)
	}

	glog.Infof("NodeStageVolume: NodeStage succeeded with VolumeId: %s, and added NodeLabel: %s", volumeId, fuseLabelKey)
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

// getRuntimeNamespacedName first checks volume context for runtime's namespace and name as a fast path.
// If not found, it takes a fallback to query API Server and to parse the PV information.
func (ns *nodeServer) getRuntimeNamespacedName(volumeContext map[string]string, volumeId string) (namespace string, name string, err error) {
	// Fast path to check namespace && name in volume context
	if len(volumeContext) != 0 {
		runtimeName, nameFound := volumeContext[common.VolumeAttrName]
		runtimeNamespace, nsFound := volumeContext[common.VolumeAttrNamespace]
		if nameFound && nsFound {
			glog.V(3).Infof("Get runtime namespace(%s) and name(%s) from volume context", runtimeNamespace, runtimeName)
			return runtimeNamespace, runtimeName, nil
		}
	}

	// Fallback: query API Server to get namespaced name
	glog.Infof("Get runtime namespace and name directly from api server with volumeId %s", volumeId)
	return volume.GetNamespacedNameByVolumeId(ns.apiReader, volumeId)
}

// getNode first checks cached node
func (ns *nodeServer) getNode() (node *v1.Node, err error) {
	// Default to allow patch stale node info
	if envVar, found := os.LookupEnv(AllowPatchStaleNodeEnv); !found || envVar == "true" {
		if ns.node != nil {
			glog.V(3).Infof("Found cached node %s", ns.node.Name)
			return ns.node, nil
		}
	}

	if node, err = ns.nodeAuthorizedClient.CoreV1().Nodes().Get(context.TODO(), ns.nodeId, metav1.GetOptions{}); err != nil {
		return nil, err
	}

	// if node, err = kubeclient.GetNode(ns.apiReader, ns.nodeId); err != nil {
	// return nil, err
	// }

	glog.V(1).Infof("Got node %s from api server", node.Name)
	ns.node = node
	return ns.node, nil
}

func (ns *nodeServer) patchNodeWithLabel(node *v1.Node, labelsToModify common.LabelsToModify) error {
	labels := labelsToModify.GetLabels()
	labelValuePair := map[string]interface{}{}

	for _, labelToModify := range labels {
		operationType := labelToModify.GetOperationType()
		labelToModifyKey := labelToModify.GetLabelKey()
		labelToModifyValue := labelToModify.GetLabelValue()

		switch operationType {
		case common.AddLabel, common.UpdateLabel:
			labelValuePair[labelToModifyKey] = labelToModifyValue
		case common.DeleteLabel:
			labelValuePair[labelToModifyKey] = nil
		default:
			err := fmt.Errorf("fail to update the label due to the wrong operation: %s", operationType)
			return err
		}
	}

	metadata := map[string]interface{}{
		"metadata": map[string]interface{}{
			"labels": labelValuePair,
		},
	}

	patchByteData, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	_, err = ns.nodeAuthorizedClient.CoreV1().Nodes().Patch(context.TODO(), node.Name, types.StrategicMergePatchType, patchByteData, metav1.PatchOptions{})
	if err != nil {
		return err
	}

	return nil
}

func checkMountInUse(volumeName string) (bool, error) {
	var inUse bool
	glog.Infof("Try to check if the volume %s is being used", volumeName)
	if volumeName == "" {
		return inUse, errors.New("volumeName is not specified")
	}

	// TODO: refer to https://github.com/kubernetes-sigs/alibaba-cloud-csi-driver/blob/4fcb743220371de82d556ab0b67b08440b04a218/pkg/oss/utils.go#L72
	// for a better implementation
	command, err := cmdguard.Command("/usr/local/bin/check_bind_mounts.sh", volumeName)
	if err != nil {
		return inUse, err
	}
	glog.Infoln(command)

	stdoutStderr, err := command.CombinedOutput()
	glog.Infoln(string(stdoutStderr))

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if waitStatus, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				exitStatus := waitStatus.ExitStatus()
				if exitStatus == 1 {
					// grep not found any mount entry
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
	}

	return inUse, err
}

// cleanUpBrokenMountPoint stats the given mountPoint and umounts it if it's broken mount point(i.e. Stat with errNo 107[Transport Endpoint is not Connected]).
func cleanUpBrokenMountPoint(mountPoint string) error {
	_, err := os.Stat(mountPoint)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		if pathErr, ok := err.(*os.PathError); ok {
			if errNo, ok := pathErr.Err.(syscall.Errno); ok {
				if errNo == syscall.ENOTCONN {
					mounter := mount.New(mountPoint)
					if err := mounter.Unmount(mountPoint); err != nil {
						return errors.Wrapf(mounter.Unmount(mountPoint), "failed to unmount %s", mountPoint)
					}
					glog.Infof("Found broken mount point %s, successfully umounted it", mountPoint)
					return nil
				}
			}
		}

		return errors.Wrapf(err, "failed to os.Stat(%s)", mountPoint)
	}

	return nil
}

func (ns *nodeServer) prepareSessMgr(workDir string) error {
	sessMgrLabelKey := common.SessMgrNodeSelectorKey
	var labelsToModify common.LabelsToModify
	labelsToModify.Add(sessMgrLabelKey, "true")

	node, err := ns.getNode()
	if err != nil {
		return errors.Wrapf(err, "can't get node %s", ns.nodeId)
	}

	// _, err = utils.ChangeNodeLabelWithPatchMode(ns.client, node, labelsToModify)
	err = ns.patchNodeWithLabel(node, labelsToModify)
	if err != nil {
		return errors.Wrapf(err, "error when patching labels on node %s", ns.nodeId)
	}

	// check sessmgrd.sock file existence
	sessMgrSockFilePath := filepath.Join(workDir, common.SessMgrSockFile)
	glog.Infof("Checking existence of file %s", sessMgrSockFilePath)
	retryLimit := 30
	var i int
	for i = 0; i < retryLimit; i++ {
		if _, err := os.Stat(sessMgrSockFilePath); err == nil {
			break
		}

		// err != nil
		if !os.IsNotExist(err) {
			glog.Errorf("fail to os.Stat sessmgr socket file %s", sessMgrSockFilePath)
		}
		time.Sleep(1 * time.Second)
	}

	if i >= retryLimit {
		return errors.New("timeout waiting for SessMgr Pod to be ready")
	}

	return nil
}

// useSymlink for nodePublishVolume if environment variable has been set or pv has attribute
func useSymlink(req *csi.NodePublishVolumeRequest) bool {
	return os.Getenv("NODEPUBLISH_METHOD") == common.NodePublishMethodSymlink || req.GetVolumeContext()[common.NodePublishMethod] == common.NodePublishMethodSymlink
}

// isLikelyNeedUnmount checks if path is likely a mount point that needs to be unmount.
// NOTE: isLikelyNeedUnmount relies on the result of mounter.IsLikelyNotMountPoint so it may not properly detect bind mounts in Linux.
func isLikelyNeedUnmount(mounter mount.Interface, path string) (needUnmount bool, err error) {
	notMount, err := mounter.IsLikelyNotMountPoint(path)
	if err != nil {
		if os.IsNotExist(err) {
			glog.V(3).Infof("NodeUnpublishVolume: targetPath %s has been cleaned up, so it doesn't need to be unmounted", path)
			return false, nil
		}

		if mount.IsCorruptedMnt(err) {
			// A corrupted path needs unmount
			glog.V(3).Infof("NodeUnpublishVolume: detected corrupted mountpoint on path %s with error %v", path, err)
			return true, nil
		}

		// unexpected errors
		return false, err
	}

	if !notMount {
		return true, nil
	}

	return false, nil
}
