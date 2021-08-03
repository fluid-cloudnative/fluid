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
	dockerapi "github.com/docker/docker/api/types"
	dockercontainer "github.com/docker/docker/api/types/container"
	dockerstrslice "github.com/docker/docker/api/types/strslice"
	dockerclient "github.com/docker/docker/client"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog"
	"os"
	"os/exec"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/common"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
	"github.com/miekg/dns"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/utils/mount"
)

const AlluxioFuseImage = "registry.aliyuncs.com/alluxio/alluxio-fuse:release-2.5.0-2-SNAPSHOT-52ad95c"
const DefaultDNSConfig = "/etc/resolv.conf"
const DNSNdotsDefault = 1
const DNSTimeoutDefault = 5
const DNSAttemptsDefault = 2

var clusterDns *dns.ClientConfig

func init() {
	var err error
	clusterDns, err = dns.ClientConfigFromFile(DefaultDNSConfig)
	if err != nil {
		klog.Errorf("got err %v when getting dns config", err)
		os.Exit(1)
	}
}

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
	cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv)
	if err != nil {
		return nil, errors.Wrap(err, "Can't new docker client")
	}

	namespacedName := strings.Split(req.GetVolumeId(), "-")
	glog.Infof("Making container run config with namespace: %s and name: %s", namespacedName[0], namespacedName[1])
	containerName := fmt.Sprintf("%s-%s-fuse", namespacedName[0], namespacedName[1])

	containerJson, err := cli.ContainerInspect(ctx, containerName)
	if err != nil {
		if dockerclient.IsErrNotFound(err) {
			return &csi.NodeUnstageVolumeResponse{}, nil
		}
		return nil, err
	}

	running := containerJson.State.Running
	timeout := 30 * time.Second
	if running {
		err = cli.ContainerStop(ctx, containerName, &timeout)
		if err != nil {
			return nil, err
		}
	}

	if err = cli.ContainerRemove(ctx, containerName, dockerapi.ContainerRemoveOptions{}); err != nil {
		return nil, err
	}

	// TODO: Extract a func here
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		var node v1.Node
		key := types.NamespacedName{
			Namespace: "",
			Name:      ns.nodeId,
		}
		if err := ns.client.Get(context.TODO(), key, &node); err != nil {
			return err
		}

		nodeToUpdate := node.DeepCopy()
		fuseLabelName := common.LabelAnnotationFusePrefix + namespacedName[0] + "-" + namespacedName[1]
		delete(nodeToUpdate.Labels, fuseLabelName)

		if !reflect.DeepEqual(node, nodeToUpdate) {
			return ns.client.Update(context.TODO(), nodeToUpdate)
		} else {
			klog.Infof("Do nothing because no label needs to be removed on node %s (label: %s)", ns.nodeId, fuseLabelName)
		}

		return nil
	})

	if err != nil {
		klog.Errorf("got err %v when labeling the node %s", err, ns.nodeId)
		return nil, errors.Wrap(err, "can't label the node")
	}

	return &csi.NodeUnstageVolumeResponse{}, nil
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

func (ns *nodeServer) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	glog.Infof("NodeStageVolume: try to start a FUSE container: %v", req)
	cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv)
	if err != nil {
		return nil, errors.Wrap(err, "Can't new docker client")
	}

	namespacedName := strings.Split(req.GetVolumeId(), "-")
	glog.Infof("Making container run config with namespace: %s and name: %s", namespacedName[0], namespacedName[1])
	containerName := fmt.Sprintf("%s-%s-fuse", namespacedName[0], namespacedName[1])

	// TODO check existence
	containerJson, err := cli.ContainerInspect(ctx, containerName)
	var running bool
	if err != nil {
		if !dockerclient.IsErrNotFound(err) {
			return nil, errors.Wrap(err, fmt.Sprintf("Can't check existence of the container"))
		}

		_, err = cli.ImagePull(ctx, AlluxioFuseImage, dockerapi.ImagePullOptions{})
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Can't pull image(%s)", AlluxioFuseImage))
		}

		//_, err = cli.ContainerInspect()

		containerConfig, hostConfig, err := ns.makeContainerRunConfig(namespacedName[0], namespacedName[1])
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Can't make container run config"))
		}

		glog.V(1).Infof(">>>>>> container config, %v", containerConfig)

		// We don't need the response because we identify the container with unique container name
		_, err = cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, containerName)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Can't create container, runConfig: %v", containerConfig))
		}
	} else {
		running = containerJson.State.Running
	}

	if !running {
		if err := cli.ContainerStart(ctx, containerName, dockerapi.ContainerStartOptions{}); err != nil {
			return nil, errors.Wrap(err, "Can't start container")
		}
	}

	// TODO: Extract a func
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		var node v1.Node
		key := types.NamespacedName{
			Namespace: "",
			Name:      ns.nodeId,
		}
		if err := ns.client.Get(context.TODO(), key, &node); err != nil {
			return err
		}

		nodeToUpdate := node.DeepCopy()
		fuseLabelName := common.LabelAnnotationFusePrefix + namespacedName[0] + "-" + namespacedName[1]
		nodeToUpdate.Labels[fuseLabelName] = "true"

		if !reflect.DeepEqual(node, nodeToUpdate) {
			err = ns.client.Update(context.TODO(), nodeToUpdate)
			if err != nil {
				return err
			}
		} else {
			klog.V(3).Infof("Do nothing because the node is already labeled (label: %s)", fuseLabelName)
		}

		return nil
	})

	if err != nil {
		klog.Errorf("got err %v when labeling the node %s", err, ns.nodeId)
		return nil, errors.Wrap(err, "can't label the node")
	}

	return &csi.NodeStageVolumeResponse{}, nil
}

func (ns *nodeServer) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (ns *nodeServer) makeContainerRunConfig(namespace, name string) (*dockercontainer.Config, *dockercontainer.HostConfig, error) {
	// 1. Get Fuse DaemonSet
	fuseDaemonSetName := name + "-fuse"
	daemonSet := &appsv1.DaemonSet{}
	err := ns.client.Get(context.TODO(), types.NamespacedName{
		Name:      fuseDaemonSetName,
		Namespace: namespace,
	}, daemonSet)

	if err != nil {
		return nil, nil, err
	}

	// 2. make envs
	containerToStart := daemonSet.Spec.Template.Spec.Containers[0]
	envs, err := ns.makeEnvironmentVariables(namespace, &containerToStart)
	if err != nil {
		return nil, nil, err
	}

	// 3. make mounts
	binds, err := ns.makeMounts(&daemonSet.Spec.Template.Spec, &containerToStart)
	if err != nil {
		return nil, nil, err
	}

	// 4. make dns configs
	dnsServers, dnsSearch, dnsOpts := ns.makeDNS(namespace)

	// 5. make capabilities
	privilege, capAdd, capDrop := ns.makeCapabilities(&daemonSet.Spec.Template.Spec, &containerToStart)

	return &dockercontainer.Config{
			Env:        envs,
			Image:      containerToStart.Image,
			Entrypoint: dockerstrslice.StrSlice(containerToStart.Command),
			Cmd:        dockerstrslice.StrSlice(containerToStart.Args),
			WorkingDir: containerToStart.WorkingDir,
			OpenStdin:  containerToStart.Stdin,
			StdinOnce:  containerToStart.StdinOnce,
			Tty:        containerToStart.TTY,
			User:       "root",
			Healthcheck: &dockercontainer.HealthConfig{
				Test: []string{"NONE"},
			},
		}, &dockercontainer.HostConfig{
			Binds: binds,
			RestartPolicy: dockercontainer.RestartPolicy{
				Name: "always",
			},
			DNS:        dnsServers,
			DNSSearch:  dnsSearch,
			DNSOptions: dnsOpts,
			Privileged: privilege,
			CapAdd:     capAdd,
			CapDrop:    capDrop,
		}, nil
}

func (ns *nodeServer) makeEnvironmentVariables(namespace string, container *v1.Container) ([]string, error) {
	var result []string
	var err error
	var (
		configMaps = make(map[string]*v1.ConfigMap)
		//secrets = make(map[string]*v1.Secret)
		tmpEnv = make(map[string]string)
	)

	for _, envFrom := range container.EnvFrom {
		switch {
		case envFrom.ConfigMapRef != nil:
			cm := envFrom.ConfigMapRef
			name := cm.Name
			configMap, ok := configMaps[name]
			if !ok {
				if ns.client == nil {
					return result, fmt.Errorf("couldn't get configMap %v/%v, no kubeClient defined", namespace, name)
				}
				optional := cm.Optional != nil && *cm.Optional
				configMap = &v1.ConfigMap{}
				err = ns.client.Get(context.TODO(), types.NamespacedName{
					Namespace: namespace,
					Name:      name,
				}, configMap)

				if err != nil {
					if apierrs.IsNotFound(err) && optional {
						continue
					}
					return result, err
				}
				configMaps[name] = configMap
			}

			for k, v := range configMap.Data {
				if len(envFrom.Prefix) > 0 {
					k = envFrom.Prefix + k
				}
				tmpEnv[k] = v
			}
		}
	}

	for _, envVar := range container.Env {
		runtimeVal := envVar.Value
		if runtimeVal != "" {
			tmpEnv[envVar.Name] = runtimeVal
		} else if envVar.ValueFrom != nil {
			// Currently we ignore such env for PoC
			continue
		}
	}

	for k, v := range tmpEnv {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}

	return result, nil
}

func (ns *nodeServer) makeMounts(podSpec *v1.PodSpec, container *v1.Container) ([]string, error) {
	var result []string

	volumeMap := make(map[string]v1.Volume)
	for _, vol := range podSpec.Volumes {
		volumeMap[vol.Name] = vol
	}

	for _, volumeMount := range container.VolumeMounts {
		if vol, ok := volumeMap[volumeMount.Name]; !ok {
			continue
		} else {
			if vol.HostPath == nil {
				continue
			} else {
				var attrs []string
				if volumeMount.ReadOnly {
					attrs = append(attrs, "ro")
				}

				if volumeMount.MountPropagation != nil {
					switch *volumeMount.MountPropagation {
					case v1.MountPropagationNone:
						//noop, private is default
					case v1.MountPropagationBidirectional:
						attrs = append(attrs, "rshared")
					case v1.MountPropagationHostToContainer:
						attrs = append(attrs, "rslave")
					default:
						glog.Warningf("unknown propagation mode for hostPath %q", vol.HostPath.Path)
					}
				}

				bind := fmt.Sprintf("%s:%s", vol.HostPath.Path, volumeMount.MountPath)
				if len(attrs) > 0 {
					bind = fmt.Sprintf("%s:%s", bind, strings.Join(attrs, ","))
				}
				result = append(result, bind)
			}
		}
	}
	return result, nil
}

func (ns *nodeServer) makeDNS(namespace string) (dnsServer, dnsSearch, dnsOptions []string) {
	dnsServer = clusterDns.Servers

	dnsSearch = clusterDns.Search
	// Transform service discovery according to the namespace of the pod
	for idx := range dnsSearch {
		if "fluid-system.svc.cluster.local" == dnsSearch[idx] {
			dnsSearch[idx] = strings.ReplaceAll(dnsSearch[idx], "fluid-system", namespace)
		}
	}

	// We ignore all the default values in dns options
	if clusterDns.Ndots != DNSNdotsDefault {
		dnsOptions = append(dnsOptions, fmt.Sprintf("ndots:%d", clusterDns.Ndots))
	}

	if clusterDns.Attempts != DNSAttemptsDefault {
		dnsOptions = append(dnsOptions, fmt.Sprintf("attempts:%d", clusterDns.Attempts))
	}

	if clusterDns.Timeout != DNSTimeoutDefault {
		dnsOptions = append(dnsOptions, fmt.Sprintf("timeout:%d", clusterDns.Timeout))
	}

	return
}

func (ns *nodeServer) makeCapabilities(podSpec *v1.PodSpec, container *v1.Container) (privileged bool, capAdd, capDrop []string) {
	privileged = *container.SecurityContext.Privileged

	for _, capability := range container.SecurityContext.Capabilities.Add {
		capAdd = append(capAdd, string(capability))
	}

	for _, capability := range container.SecurityContext.Capabilities.Drop {
		capDrop = append(capDrop, string(capability))
	}

	return
}

// TODO: Change users
//func (ns *nodeServer) makeUsers(container *v1.Container) (user string) {
//
//}
