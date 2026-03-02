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

package recover

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/dataset/volume"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubelet"
	"github.com/fluid-cloudnative/fluid/pkg/utils/mountinfo"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	k8sexec "k8s.io/utils/exec"
	"k8s.io/utils/mount"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	defaultKubeletTimeout          = 10
	defaultFuseRecoveryPeriod      = 5 * time.Second
	defaultRecoverWarningThreshold = 50
	serviceAccountTokenFile        = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	FuseRecoveryPeriod             = "RECOVER_FUSE_PERIOD"
	RecoverWarningThreshold        = "RECOVER_WARNING_THRESHOLD"
)

var _ manager.Runnable = &FuseRecover{}

type FuseRecover struct {
	mount.SafeFormatAndMount
	KubeClient client.Client
	ApiReader  client.Reader
	// KubeletClient *kubelet.KubeletClient
	Recorder record.EventRecorder

	recoverFusePeriod       time.Duration
	recoverWarningThreshold int

	locks *utils.VolumeLocks
}

func initializeKubeletClient() (*kubelet.KubeletClient, error) {
	// get CSI sa token
	tokenByte, err := os.ReadFile(serviceAccountTokenFile)
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

func NewFuseRecover(kubeClient client.Client, recorder record.EventRecorder, apiReader client.Reader, locks *utils.VolumeLocks) (*FuseRecover, error) {
	glog.V(3).Infoln("start csi recover")
	mountRoot, err := utils.GetMountRoot()
	if err != nil {
		return nil, errors.Wrap(err, "got err when getting mount root")
	}
	glog.V(3).Infof("Get mount root: %s", mountRoot)

	if err != nil {
		return nil, errors.Wrap(err, "got error when creating kubelet client")
	}

	recoverFusePeriod := utils.GetDurationValueFromEnv(FuseRecoveryPeriod, defaultFuseRecoveryPeriod)
	recoverWarningThreshold, found := utils.GetIntValueFromEnv(RecoverWarningThreshold)
	if !found {
		recoverWarningThreshold = defaultRecoverWarningThreshold
	}
	return &FuseRecover{
		SafeFormatAndMount: mount.SafeFormatAndMount{
			Interface: mount.New(""),
			Exec:      k8sexec.New(),
		},
		KubeClient:              kubeClient,
		ApiReader:               apiReader,
		Recorder:                recorder,
		recoverFusePeriod:       recoverFusePeriod,
		recoverWarningThreshold: recoverWarningThreshold,
		locks:                   locks,
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
	go wait.Until(r.runOnce, r.recoverFusePeriod, stopCh)
	<-stopCh
	glog.V(3).Info("Shutdown CSI recover.")
}

func (r *FuseRecover) runOnce() {
	r.recover()
}

func (r *FuseRecover) recover() {
	brokenMounts, err := mountinfo.GetBrokenMountPoints()
	if err != nil {
		// Error: This is an unexpected failure reading mount info - not a recoverable mount issue
		glog.Errorf("FuseRecovery: failed to get broken mount points: %v", err)
		return
	}

	if len(brokenMounts) > 0 {
		// Info: State transition - detected broken mounts that need recovery
		glog.V(3).Infof("FuseRecovery: detected %d broken mount point(s) requiring recovery", len(brokenMounts))
	}

	for _, point := range brokenMounts {
		r.doRecover(point)
	}
}

func (r *FuseRecover) recoverBrokenMount(point mountinfo.MountPoint) (err error) {
	// recovery for each bind mount path
	mountOption := []string{"bind"}
	if point.ReadOnly {
		mountOption = append(mountOption, "ro")
	}

	// Info: Attempting recovery action
	glog.V(3).Infof("FuseRecovery: attempting bind mount, source=%s mountPath=%s options=%v", point.SourcePath, point.MountPath, mountOption)
	if err := r.Mount(point.SourcePath, point.MountPath, "none", mountOption); err != nil {
		// Warning: Mount failure is recoverable - will retry on next cycle
		glog.Warningf("FuseRecovery: bind mount failed, mountPath=%s source=%s error=%v", point.MountPath, point.SourcePath, err)
	}
	return
}

// check mountpoint count
// umount duplicate mountpoint util 1 avoiding very large mountinfo file.
// don't umount all item, 'mountPropagation' will lose efficacy.
func (r *FuseRecover) umountDuplicate(point mountinfo.MountPoint) {
	for i := point.Count; i > 1; i-- {
		// Info: Cleanup action - removing duplicate mount
		glog.V(3).Infof("FuseRecovery: unmounting duplicate, mountPath=%s remainingCount=%d", point.MountPath, i)
		if err := r.Unmount(point.MountPath); err != nil {
			// Warning: Unmount failure is recoverable - will retry
			glog.Warningf("FuseRecovery: unmount failed, mountPath=%s error=%v", point.MountPath, err)
		}
	}
}

func (r *FuseRecover) eventRecord(point mountinfo.MountPoint, eventType, eventReason string) {
	namespacedName := point.NamespacedDatasetName
	strs := strings.Split(namespacedName, "-")
	if len(strs) < 2 {
		glog.V(3).Infof("can't parse dataset from namespacedName: %s", namespacedName)
		return
	}
	namespace, datasetName, err := volume.GetNamespacedNameByVolumeId(r.ApiReader, namespacedName)
	if err != nil {
		// Error: API failure - cannot record event without dataset reference
		glog.Errorf("FuseRecovery: failed to get namespace/name by volumeId=%s error=%v", namespacedName, err)
		return
	}

	dataset, err := utils.GetDataset(r.KubeClient, datasetName, namespace)
	if err != nil {
		// Error: API failure - dataset lookup failed
		glog.Errorf("FuseRecovery: failed to get dataset, name=%s namespace=%s error=%v", datasetName, namespace, err)
		return
	}
	glog.V(4).Infof("record to dataset: %s, namespace: %s", dataset.Name, dataset.Namespace)
	switch eventReason {
	case common.FuseRecoverSucceed:
		r.Recorder.Eventf(dataset, eventType, eventReason, "Fuse recover %s succeed", point.MountPath)
	case common.FuseRecoverFailed:
		r.Recorder.Eventf(dataset, eventType, eventReason, "Fuse recover %s failed", point.MountPath)
	case common.FuseUmountDuplicate:
		r.Recorder.Eventf(dataset, eventType, eventReason, "Mountpoint %s has been mounted %v times, unmount duplicate mountpoint to avoid large /proc/self/mountinfo file, this may potential make data access connection broken", point.MountPath, point.Count)
	}
}

func (r *FuseRecover) shouldRecover(mountPath string) (should bool, err error) {
	mounter := mount.New("")
	notMount, err := mounter.IsLikelyNotMountPoint(mountPath)
	if os.IsNotExist(err) || (err == nil && notMount) {
		// Perhaps the mountPath has been cleaned up in other goroutine
		return false, nil
	}
	if err != nil && !mount.IsCorruptedMnt(err) {
		// unexpected error
		return false, err
	}

	return true, nil
}

func (r *FuseRecover) doRecover(point mountinfo.MountPoint) {
	if lock := r.locks.TryAcquire(point.MountPath); !lock {
		// Info: Concurrent recovery in progress - skip to avoid race
		glog.V(4).Infof("FuseRecovery: skipping recovery, lock held by another goroutine, mountPath=%s", point.MountPath)
		return
	}
	defer r.locks.Release(point.MountPath)

	should, err := r.shouldRecover(point.MountPath)
	if err != nil {
		// Warning: Recovery check failed - may retry next cycle
		glog.Warningf("FuseRecovery: unable to determine recovery state, mountPath=%s error=%v", point.MountPath, err)
		return
	}

	if !should {
		// Info: Mount already healthy or cleaned up
		glog.V(3).Infof("FuseRecovery: skipping recovery, mount already cleaned up, mountPath=%s", point.MountPath)
		return
	}

	// Info: Starting recovery attempt
	glog.V(3).Infof("FuseRecovery: starting recovery, mountPath=%s source=%s mountCount=%d", point.MountPath, point.SourcePath, point.Count)

	// if app container restart, umount duplicate mount may lead to recover successes but can not access data
	// so we only umountDuplicate when it has mounted more than the recoverWarningThreshold
	// please refer to https://github.com/fluid-cloudnative/fluid/issues/3399 for more information
	if point.Count > r.recoverWarningThreshold {
		// Warning: Excessive mount count detected - cleanup required
		glog.Warningf("FuseRecovery: excessive mount count detected, mountPath=%s count=%d threshold=%d", point.MountPath, point.Count, r.recoverWarningThreshold)
		r.eventRecord(point, corev1.EventTypeWarning, common.FuseUmountDuplicate)
		r.umountDuplicate(point)
	}

	if err := r.recoverBrokenMount(point); err != nil {
		// Warning logged inside recoverBrokenMount, just record event
		r.eventRecord(point, corev1.EventTypeWarning, common.FuseRecoverFailed)
		return
	}

	// Info: Recovery succeeded - state transition from broken to healthy
	glog.V(3).Infof("FuseRecovery: recovery succeeded, mountPath=%s", point.MountPath)
	r.eventRecord(point, corev1.EventTypeNormal, common.FuseRecoverSucceed)
}
