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

package recover

import (
	"context"
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/dataset/volume"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubelet"
	"github.com/fluid-cloudnative/fluid/pkg/utils/mountinfo"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
	k8sexec "k8s.io/utils/exec"
	"k8s.io/utils/mount"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"strconv"
	"strings"
	"time"
)

const (
	defaultKubeletTimeout     = 10
	defaultFuseRecoveryPeriod = 5
	serviceAccountTokenFile   = "/var/run/secrets/kubernetes.io/serviceaccount/token"
)

var _ manager.Runnable = &FuseRecover{}

type FuseRecover struct {
	mount.SafeFormatAndMount
	KubeClient    client.Client
	KubeletClient *kubelet.KubeletClient
	Recorder      record.EventRecorder

	containers map[string]*containerStat // key: <containerName>-<daemonSetName>-<namespace>

	recoverFusePeriod int
}

type containerStat struct {
	name          string
	podName       string
	namespace     string
	daemonSetName string
	startAt       metav1.Time
}

func initializeKubeletClient() (*kubelet.KubeletClient, error) {
	// get CSI sa token
	tokenByte, err := ioutil.ReadFile(serviceAccountTokenFile)
	if err != nil {
		return nil, errors.Wrap(err, "in cluster mode, find token failed")
	}
	token := string(tokenByte)

	glog.V(3).Infoln("start kubelet client")
	nodeIp := os.Getenv("NODE_IP")
	kubeletClientCert := os.Getenv("KUBELET_CLIENT_CERT")
	kubeletClientKey := os.Getenv("KUBELET_CLIENT_KEY")
	var kubeletTimeout int
	if os.Getenv("KUBELET_TIMEOUT") != "" {
		if kubeletTimeout, err = strconv.Atoi(os.Getenv("KUBELET_TIMEOUT")); err != nil {
			return nil, errors.Wrap(err, "got error when parsing kubelet timeout")
		}
	} else {
		kubeletTimeout = defaultKubeletTimeout
	}
	glog.V(3).Infof("get node ip: %s", nodeIp)
	kubeletClient, err := kubelet.NewKubeletClient(&kubelet.KubeletClientConfig{
		Address: nodeIp,
		Port:    10250,
		TLSClientConfig: rest.TLSClientConfig{
			ServerName: "kubelet",
			CertFile:   kubeletClientCert,
			KeyFile:    kubeletClientKey,
		},
		BearerToken: token,
		HTTPTimeout: time.Duration(kubeletTimeout) * time.Second,
	})

	if err != nil {
		return nil, err
	}

	return kubeletClient, nil
}

func NewFuseRecover(kubeClient client.Client, recorder record.EventRecorder) (*FuseRecover, error) {
	glog.V(3).Infoln("start csi recover")
	mountRoot, err := utils.GetMountRoot()
	if err != nil {
		return nil, errors.Wrap(err, "got err when getting mount root")
	}
	glog.V(3).Infof("Get mount root: %s", mountRoot)

	if err != nil {
		return nil, errors.Wrap(err, "got error when creating kubelet client")
	}

	kubeletClient, err := initializeKubeletClient()
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize kubelet")
	}

	recoverFusePeriod := defaultFuseRecoveryPeriod
	if os.Getenv("RECOVER_FUSE_PERIOD") != "" {
		recoverFusePeriod, err = strconv.Atoi(os.Getenv("RECOVER_FUSE_PERIOD"))
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse period to int")
		}
	}
	return &FuseRecover{
		SafeFormatAndMount: mount.SafeFormatAndMount{
			Interface: mount.New(""),
			Exec:      k8sexec.New(),
		},
		KubeClient:        kubeClient,
		KubeletClient:     kubeletClient,
		Recorder:          recorder,
		containers:        make(map[string]*containerStat),
		recoverFusePeriod: recoverFusePeriod,
	}, nil
}

func (r *FuseRecover) Start(ctx context.Context) error {
	// do recovering at beginning
	// recover set containerStat in memory, it's none when start
	r.recover()
	r.run(wait.NeverStop)

	return nil
}

func (r *FuseRecover) run(stopCh <-chan struct{}) {
	go wait.Until(r.runOnce, time.Duration(r.recoverFusePeriod)*time.Second, stopCh)
	<-stopCh
	glog.V(3).Info("Shutdown CSI recover.")
}

func (r *FuseRecover) runOnce() {
	pods, err := r.KubeletClient.GetNodeRunningPods()
	glog.V(6).Info("get pods from kubelet")
	if err != nil {
		glog.Error(err)
		return
	}
	for _, pod := range pods.Items {
		glog.V(6).Infof("get pod: %s, namespace: %s", pod.Name, pod.Namespace)
		if !utils.IsFusePod(pod) {
			continue
		}
		if !podutil.IsPodReady(&pod) {
			continue
		}
		glog.V(6).Infof("get fluid fuse pod: %s, namespace: %s", pod.Name, pod.Namespace)
		if isRestarted := r.compareOrRecordContainerStat(pod); isRestarted {
			glog.V(3).Infof("fuse pod restarted: %s, namespace: %s", pod.Name, pod.Namespace)
			r.recover()
			return
		}
	}
}

func (r FuseRecover) recover() {
	brokenMounts, err := mountinfo.GetBrokenMountPoints()
	if err != nil {
		glog.Error(err)
		return
	}

	for _, point := range brokenMounts {
		glog.V(4).Infof("Get broken mount point: %v", point)
		r.umountDuplicate(point)
		if err := r.recoverBrokenMount(point); err != nil {
			r.eventRecord(point, corev1.EventTypeWarning, common.FuseRecoverFailed)
			continue
		}
		r.eventRecord(point, corev1.EventTypeNormal, common.FuseRecoverSucceed)
	}
}

func (r *FuseRecover) recoverBrokenMount(point mountinfo.MountPoint) (err error) {
	glog.V(3).Infof("Start recovery: [%s], source path: [%s]", point.MountPath, point.SourcePath)
	// recovery for each bind mount path
	mountOption := []string{"bind"}
	if point.ReadOnly {
		mountOption = append(mountOption, "ro")
	}

	glog.V(3).Infof("Start exec cmd: mount %s %s -o %v \n", point.SourcePath, point.MountPath, mountOption)
	if err := r.Mount(point.SourcePath, point.MountPath, "none", mountOption); err != nil {
		glog.Errorf("exec cmd: mount -o bind %s %s err :%v", point.SourcePath, point.MountPath, err)
	}
	return
}

// check mountpoint count
// umount duplicate mountpoint util 1 avoiding very large mountinfo file.
// don't umount all item, 'mountPropagation' will lose efficacy.
func (r *FuseRecover) umountDuplicate(point mountinfo.MountPoint) {
	for i := point.Count; i > 1; i-- {
		glog.V(3).Infof("count: %d, start exec cmd: umount %s", i, point.MountPath)
		if err := r.Unmount(point.MountPath); err != nil {
			glog.Errorf("exec cmd: umount %s err: %v", point.MountPath, err)
		}
	}
}

func (r *FuseRecover) compareOrRecordContainerStat(pod corev1.Pod) (restarted bool) {
	if pod.Status.ContainerStatuses == nil || len(pod.OwnerReferences) == 0 {
		return
	}
	var dsName string
	for _, obj := range pod.OwnerReferences {
		if obj.Kind == "DaemonSet" {
			dsName = obj.Name
		}
	}
	if dsName == "" {
		return
	}
	for _, cn := range pod.Status.ContainerStatuses {
		if cn.State.Running == nil {
			continue
		}
		key := fmt.Sprintf("%s-%s-%s", cn.Name, dsName, pod.Namespace)
		cs, ok := r.containers[key]
		if !ok {
			cs = &containerStat{
				name:          cn.Name,
				podName:       pod.Name,
				namespace:     pod.Namespace,
				daemonSetName: dsName,
				startAt:       cn.State.Running.StartedAt,
			}
			r.containers[key] = cs
			continue
		}

		if cs.startAt.Before(&cn.State.Running.StartedAt) {
			glog.Infof("Container %s of pod %s in namespace %s start time is %s, but record %s",
				cn.Name, pod.Name, pod.Namespace, cn.State.Running.StartedAt.String(), cs.startAt.String())
			r.containers[key].startAt = cn.State.Running.StartedAt
			restarted = true
			return
		}
	}
	return
}

func (r *FuseRecover) eventRecord(point mountinfo.MountPoint, eventType, eventReason string) {
	namespacedName := point.NamespacedDatasetName
	strs := strings.Split(namespacedName, "-")
	if len(strs) < 2 {
		glog.V(3).Infof("can't parse dataset from namespacedName: %s", namespacedName)
		return
	}
	namespace, datasetName, err := volume.GetNamespacedNameByVolumeId(r.KubeClient, namespacedName)
	if err != nil {
		glog.Errorf("error get namespacedName by volume id %s: %v", namespacedName, err)
		return
	}

	dataset, err := utils.GetDataset(r.KubeClient, datasetName, namespace)
	if err != nil {
		glog.Errorf("error get dataset %s namespace %s: %v", datasetName, namespace, err)
		return
	}
	glog.V(4).Infof("record to dataset: %s, namespace: %s", dataset.Name, dataset.Namespace)
	r.Recorder.Eventf(dataset, eventType, eventReason, "Fuse recover %s succeed", point.MountPath)
}
