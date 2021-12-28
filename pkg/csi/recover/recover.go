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
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/csi/mountinfo"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubelet"
	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/record"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
	k8sexec "k8s.io/utils/exec"
	"k8s.io/utils/mount"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"
)

type FuseRecover struct {
	mount.SafeFormatAndMount
	KubeClient    client.Client
	KubeletClient *kubelet.KubeletClient
	Recorder      record.EventRecorder

	containers map[string]*containerStat // key: <containerName>-<daemonSetName>-<namespace>
}

type containerStat struct {
	name          string
	podName       string
	namespace     string
	daemonSetName string
	startAt       metav1.Time
}

func NewFuseRecoder(kubeClient client.Client, kubeletClient *kubelet.KubeletClient, recorder record.EventRecorder) *FuseRecover {
	return &FuseRecover{
		SafeFormatAndMount: mount.SafeFormatAndMount{
			Interface: mount.New(""),
			Exec:      k8sexec.New(),
		},
		KubeClient:    kubeClient,
		KubeletClient: kubeletClient,
		Recorder:      recorder,
		containers:    make(map[string]*containerStat),
	}
}

func (r *FuseRecover) Run(period int, stopCh <-chan struct{}) {
	go wait.Until(r.run, time.Duration(period)*time.Second, stopCh)
	<-stopCh
	glog.V(3).Info("Shutdown CSI recover.")
}

func (r *FuseRecover) run() {
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
			r.Recover()
			return
		}
	}
}

func (r FuseRecover) Recover() {
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
	namespace, datasetName := strs[0], strs[1]

	dataset, err := utils.GetDataset(r.KubeClient, datasetName, namespace)
	if err != nil {
		glog.Errorf("error get dataset %s namespace %s: %v", datasetName, namespace, err)
		return
	}
	glog.V(4).Infof("record to dataset: %s, namespace: %s", dataset.Name, dataset.Namespace)
	r.Recorder.Eventf(dataset, eventType, eventReason, "Fuse recover %s succeed", point.MountPath)
}
